// ct-----------------------------------------
// @file      : matcher.go
// @time      : 2025/01/24 21:00
// -------------------------------------------

// ============================================

package ct

import (
	"fmt"
	"hash/fnv"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/publicsuffix"
	"recon/internal/utils/ct/models"
	"recon/internal/utils/ct/types"
)

type DomainMatcher struct {
	config      *types.DomainMatcherConfig
	trieRoot    *models.DomainTrieNode
	bloomFilter *BloomFilter
	lruCache    *LRUCache
	domains     map[string]*models.Domain
	mu          sync.RWMutex

	stats struct {
		totalMatches int64
		cacheHits    int64
		cacheMisses  int64
		bloomHits    int64
		bloomMisses  int64
		trieHits     int64
		trieMisses   int64
	}
}

func NewDomainMatcher(config *types.DomainMatcherConfig) *DomainMatcher {
	dm := &DomainMatcher{
		config:   config,
		trieRoot: models.NewDomainTrieNode(),
		domains:  make(map[string]*models.Domain),
	}

	if config.EnableBloomFilter {
		dm.bloomFilter = NewBloomFilter(config.BloomFilterSize, config.BloomFilterHash)
	}

	if config.EnableLRUCache {
		dm.lruCache = NewLRUCache(config.LRUCacheSize)
	}

	return dm
}

func (dm *DomainMatcher) AddDomain(domain string) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	domain = dm.normalizeDomain(domain)
	if domain == "" {
		return fmt.Errorf("Invalid domain: %s", domain)
	}

	if _, exists := dm.domains[domain]; exists {
		return nil
	}

	domainObj := models.NewDomain(domain)

	dm.domains[domain] = domainObj

	dm.addToTrie(domain, domainObj)

	if dm.bloomFilter != nil {
		dm.bloomFilter.Add(domain)
	}

	return nil
}

func (dm *DomainMatcher) RemoveDomain(domain string) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	domain = dm.normalizeDomain(domain)
	if domain == "" {
		return nil
	}

	delete(dm.domains, domain)

	dm.removeFromTrie(domain)

	if dm.bloomFilter != nil {
		dm.rebuildBloomFilter()
	}

	if dm.lruCache != nil {
		dm.lruCache.Clear()
	}

	return nil
}

func (dm *DomainMatcher) HasDomain(domain string) bool {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	domain = dm.normalizeDomain(domain)
	_, exists := dm.domains[domain]
	return exists
}

func (dm *DomainMatcher) GetAllDomains() []string {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	domains := make([]string, 0, len(dm.domains))
	for domain := range dm.domains {
		domains = append(domains, domain)
	}
	return domains
}

func (dm *DomainMatcher) MatchSubdomains(domains []string) []*models.DomainMatch {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	var matches []*models.DomainMatch

	for _, domain := range domains {
		domain = dm.normalizeDomain(domain)
		if domain == "" {
			continue
		}

		rootDomain := extractRootDomain(domain)
		if rootDomain == "" {
			continue
		}

		if dm.bloomFilter != nil {
			if !dm.bloomFilter.Contains(rootDomain) {
				dm.stats.bloomMisses++
				continue
			}
			dm.stats.bloomHits++
		}

		if dm.lruCache != nil {
			if _, found := dm.lruCache.Get(domain); found {
				dm.stats.cacheHits++
				continue
			}
			dm.stats.cacheMisses++
		}

		if match := dm.matchInTrie(rootDomain); match != nil {

			match.Subdomain = domain
			dm.stats.trieHits++
			dm.stats.totalMatches++

			matches = append(matches, match)

			if dm.lruCache != nil {
				dm.lruCache.Put(domain, match)
			}
		} else {
			dm.stats.trieMisses++
		}
	}

	return matches
}

func (dm *DomainMatcher) IsSubdomainOf(subdomain, rootDomain string) bool {
	subdomain = dm.normalizeDomain(subdomain)
	rootDomain = dm.normalizeDomain(rootDomain)

	if subdomain == "" || rootDomain == "" {
		return false
	}

	if subdomain == rootDomain {
		return false
	}

	if !strings.HasSuffix(subdomain, "."+rootDomain) {
		return false
	}

	subParts := strings.Split(subdomain, ".")
	rootParts := strings.Split(rootDomain, ".")

	return len(subParts) > len(rootParts)
}

func (dm *DomainMatcher) GetStats() *models.DomainMatcherStats {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	stats := &models.DomainMatcherStats{
		DomainStats: models.DomainStats{
			TotalDomains:    int64(len(dm.domains)),
			WildcardDomains: 0,
		},
	}

	for _, domain := range dm.domains {
		if domain.IsWildcard {
			stats.WildcardDomains++
		}
	}

	if dm.bloomFilter != nil {
		stats.BloomFilterStats = models.BloomFilterStats{
			Size:      dm.bloomFilter.size,
			HashCount: dm.bloomFilter.hashCount,
			ItemCount: dm.bloomFilter.itemCount,
		}
	}

	stats.TrieStats = dm.getTrieStats()

	if dm.lruCache != nil {
		stats.CacheStats = models.CacheStats{
			Size:      dm.lruCache.capacity,
			ItemCount: len(dm.lruCache.items),
			HitCount:  dm.stats.cacheHits,
			MissCount: dm.stats.cacheMisses,
		}
		if dm.stats.cacheHits+dm.stats.cacheMisses > 0 {
			stats.CacheStats.HitRate = float64(dm.stats.cacheHits) / float64(dm.stats.cacheHits+dm.stats.cacheMisses)
		}
	}

	return stats
}

func (dm *DomainMatcher) Clear() error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	dm.domains = make(map[string]*models.Domain)
	dm.trieRoot = models.NewDomainTrieNode()

	if dm.bloomFilter != nil {
		dm.bloomFilter.Clear()
	}

	if dm.lruCache != nil {
		dm.lruCache.Clear()
	}

	dm.stats = struct {
		totalMatches int64
		cacheHits    int64
		cacheMisses  int64
		bloomHits    int64
		bloomMisses  int64
		trieHits     int64
		trieMisses   int64
	}{}

	return nil
}

func (dm *DomainMatcher) normalizeDomain(domain string) string {
	return models.NormalizeDomain(domain, dm.config.CaseSensitive)
}

func (dm *DomainMatcher) addToTrie(domain string, domainObj *models.Domain) {
	node := dm.trieRoot
	runes := []rune(domain)

	for _, r := range runes {
		if node.Children[r] == nil {
			node.Children[r] = models.NewDomainTrieNode()
		}
		node = node.Children[r]
	}

	node.IsEnd = true
	node.Domain = domain
	node.AddedAt = time.Now()
	node.DomainObj = domainObj
}

func (dm *DomainMatcher) removeFromTrie(domain string) {

	dm.rebuildTrie()
}

func (dm *DomainMatcher) rebuildTrie() {
	dm.trieRoot = models.NewDomainTrieNode()
	for domain, domainObj := range dm.domains {
		dm.addToTrie(domain, domainObj)
	}
}

func (dm *DomainMatcher) rebuildBloomFilter() {
	if dm.bloomFilter == nil {
		return
	}

	dm.bloomFilter.Clear()
	for domain := range dm.domains {
		dm.bloomFilter.Add(domain)
	}
}

func (dm *DomainMatcher) matchInTrie(rootDomain string) *models.DomainMatch {

	if node := dm.searchTrie(rootDomain); node != nil && node.IsEnd {
		node.HitCount++
		return &models.DomainMatch{
			Subdomain:  "",
			RootDomain: rootDomain,
			MatchedAt:  time.Now(),
			MatchType:  "root_domain",
			Confidence: 1.0,
			HitCount:   node.HitCount,
		}
	}

	return nil
}

func (dm *DomainMatcher) searchTrie(domain string) *models.DomainTrieNode {
	node := dm.trieRoot
	runes := []rune(domain)

	for _, r := range runes {
		if node.Children[r] == nil {
			return nil
		}
		node = node.Children[r]
	}

	return node
}

func (dm *DomainMatcher) getTrieStats() models.TrieStats {
	stats := models.TrieStats{}
	dm.traverseTrie(dm.trieRoot, 0, &stats)
	return stats
}

func (dm *DomainMatcher) traverseTrie(node *models.DomainTrieNode, depth int, stats *models.TrieStats) {
	stats.NodeCount++

	if depth > stats.MaxDepth {
		stats.MaxDepth = depth
	}

	if node.IsEnd {
		stats.DomainCount++
		stats.TotalHitCount += node.HitCount
	}

	for _, child := range node.Children {
		dm.traverseTrie(child, depth+1, stats)
	}
}

type BloomFilter struct {
	bitset    []bool
	hashes    []uint
	size      uint
	hashCount uint
	itemCount uint64
}

func NewBloomFilter(size, hashCount uint) *BloomFilter {
	return &BloomFilter{
		bitset:    make([]bool, size),
		hashes:    make([]uint, hashCount),
		size:      size,
		hashCount: hashCount,
	}
}

func (bf *BloomFilter) Add(item string) {
	for i := uint(0); i < bf.hashCount; i++ {
		hash := bf.hash(item, i)
		bf.bitset[hash%bf.size] = true
	}
	bf.itemCount++
}

func (bf *BloomFilter) Contains(item string) bool {
	for i := uint(0); i < bf.hashCount; i++ {
		hash := bf.hash(item, i)
		if !bf.bitset[hash%bf.size] {
			return false
		}
	}
	return true
}

func (bf *BloomFilter) Clear() {
	bf.bitset = make([]bool, bf.size)
	bf.itemCount = 0
}

func (bf *BloomFilter) hash(item string, seed uint) uint {

	h1 := fnv.New32a()
	h1.Write([]byte(item))
	hash1 := uint64(h1.Sum32())

	h2 := fnv.New32()
	h2.Write([]byte(item))
	hash2 := uint64(h2.Sum32())

	hash := hash1 + uint64(seed)*hash2
	return uint(hash % uint64(bf.size))
}

func extractRootDomain(domain string) string {
	domain = strings.ToLower(strings.TrimSpace(domain))
	if domain == "" {
		return ""
	}

	suffix, icann := publicsuffix.PublicSuffix(domain)
	if suffix == "" || !icann {

		parts := strings.Split(domain, ".")
		if len(parts) >= 2 {
			return strings.Join(parts[len(parts)-2:], ".")
		}
		return domain
	}

	root, err := publicsuffix.EffectiveTLDPlusOne(domain)
	if err != nil {

		parts := strings.Split(domain, ".")
		if len(parts) >= 2 {
			return strings.Join(parts[len(parts)-2:], ".")
		}
		return domain
	}

	return root
}

type LRUCache struct {
	capacity int
	items    map[string]interface{}
	keys     []string
}

func NewLRUCache(capacity int) *LRUCache {
	return &LRUCache{
		capacity: capacity,
		items:    make(map[string]interface{}),
		keys:     make([]string, 0, capacity),
	}
}

func (c *LRUCache) Get(key string) (interface{}, bool) {
	if value, exists := c.items[key]; exists {

		c.moveToEnd(key)
		return value, true
	}
	return nil, false
}

func (c *LRUCache) Put(key string, value interface{}) {
	if _, exists := c.items[key]; exists {
		c.items[key] = value
		c.moveToEnd(key)
		return
	}

	if len(c.items) >= c.capacity {

		oldest := c.keys[0]
		delete(c.items, oldest)
		c.keys = c.keys[1:]
	}

	c.items[key] = value
	c.keys = append(c.keys, key)
}

func (c *LRUCache) Clear() {
	c.items = make(map[string]interface{})
	c.keys = c.keys[:0]
}

func (c *LRUCache) moveToEnd(key string) {
	for i, k := range c.keys {
		if k == key {
			c.keys = append(c.keys[:i], c.keys[i+1:]...)
			c.keys = append(c.keys, key)
			break
		}
	}
}
