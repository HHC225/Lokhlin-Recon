// ct/models-----------------------------------
// @file      : domain.go
// @time      : 2025/01/24 21:00
// -------------------------------------------

// ============================================

package models

import (
	"fmt"
	"strings"
	"time"
	"unicode"
)

type Domain struct {
	Name       string    `json:"name"`
	IsWildcard bool      `json:"is_wildcard"`
	RootDomain string    `json:"root_domain"`
	Subdomains []string  `json:"subdomains"`
	AddedAt    time.Time `json:"added_at"`
	LastSeen   time.Time `json:"last_seen"`
	CertCount  int64     `json:"cert_count"`
	HitCount   int64     `json:"hit_count"`
}

func NewDomain(name string) *Domain {
	now := time.Now()
	return &Domain{
		Name:     name,
		AddedAt:  now,
		LastSeen: now,
	}
}

func (d *Domain) IsSubdomainOf(rootDomain string) bool {
	if d.Name == rootDomain {
		return false
	}

	if !strings.HasSuffix(d.Name, "."+rootDomain) {
		return false
	}

	subParts := strings.Split(d.Name, ".")
	rootParts := strings.Split(rootDomain, ".")

	return len(subParts) > len(rootParts)
}

func (d *Domain) UpdateLastSeen() {
	d.LastSeen = time.Now()
	d.HitCount++
}

func (d *Domain) IncrementCertCount() {
	d.CertCount++
}

type DomainMatch struct {
	Subdomain   string    `json:"subdomain"`
	RootDomain  string    `json:"root_domain"`
	Certificate *CertInfo `json:"certificate"`
	MatchedAt   time.Time `json:"matched_at"`
	MatchType   string    `json:"match_type"`
	Confidence  float64   `json:"confidence"`
	HitCount    int64     `json:"hit_count"`
}

func NewDomainMatch(subdomain, rootDomain string, cert *CertInfo) *DomainMatch {
	return &DomainMatch{
		Subdomain:   subdomain,
		RootDomain:  rootDomain,
		Certificate: cert,
		MatchedAt:   time.Now(),
		MatchType:   "prefix",
		Confidence:  0.9,
		HitCount:    1,
	}
}

type DomainTrieNode struct {
	Children  map[rune]*DomainTrieNode `json:"-"`
	IsEnd     bool                     `json:"is_end"`
	Domain    string                   `json:"domain"`
	AddedAt   time.Time                `json:"added_at"`
	HitCount  int64                    `json:"hit_count"`
	DomainObj *Domain                  `json:"-"`
}

func NewDomainTrieNode() *DomainTrieNode {
	return &DomainTrieNode{
		Children: make(map[rune]*DomainTrieNode),
		IsEnd:    false,
	}
}

type DomainStats struct {
	TotalDomains      int64   `json:"total_domains"`
	UniqueSubdomains  int64   `json:"unique_subdomains"`
	NewDiscoveries    int64   `json:"new_discoveries"`
	RootDomainCount   int64   `json:"root_domain_count"`
	WildcardDomains   int64   `json:"wildcard_domains"`
	TotalMatches      int64   `json:"total_matches"`
	AverageConfidence float64 `json:"average_confidence"`
	CacheHitRate      float64 `json:"cache_hit_rate"`
}

type DomainMatcherStats struct {
	DomainStats
	BloomFilterStats BloomFilterStats `json:"bloom_filter_stats"`
	TrieStats        TrieStats        `json:"trie_stats"`
	CacheStats       CacheStats       `json:"cache_stats"`
}

type BloomFilterStats struct {
	Size              uint    `json:"size"`
	HashCount         uint    `json:"hash_count"`
	ItemCount         uint64  `json:"item_count"`
	FalsePositiveRate float64 `json:"false_positive_rate"`
}

type TrieStats struct {
	NodeCount     int64 `json:"node_count"`
	MaxDepth      int   `json:"max_depth"`
	DomainCount   int64 `json:"domain_count"`
	TotalHitCount int64 `json:"total_hit_count"`
}

type CacheStats struct {
	Size      int     `json:"size"`
	ItemCount int     `json:"item_count"`
	HitCount  int64   `json:"hit_count"`
	MissCount int64   `json:"miss_count"`
	HitRate   float64 `json:"hit_rate"`
}

type DomainMatcherInterface interface {
	AddDomain(domain string) error
	RemoveDomain(domain string) error
	HasDomain(domain string) bool
	GetAllDomains() []string

	MatchSubdomains(domains []string) []*DomainMatch
	IsSubdomainOf(subdomain, rootDomain string) bool

	GetStats() *DomainMatcherStats
	Clear() error
}

func NormalizeDomain(domain string, caseSensitive bool) string {
	if domain == "" {
		return ""
	}

	domain = strings.TrimSpace(domain)

	if !caseSensitive {
		domain = strings.ToLower(domain)
	}

	domain = strings.TrimSuffix(domain, ".")

	return domain
}

func ValidateDomain(domain string, maxLength int) error {
	if domain == "" {
		return fmt.Errorf("Domain is required")
	}

	if len(domain) > maxLength {
		return fmt.Errorf("Domain exceeds maximum length %d", maxLength)
	}

	for _, r := range domain {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '-' && r != '.' && r != '*' {
			return fmt.Errorf("Domain contains an invalid character: %c", r)
		}
	}

	return nil
}
