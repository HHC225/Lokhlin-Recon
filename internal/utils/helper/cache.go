// Package helper -----------------------------
// @file      : cache.go
// @time      : 2025/1/16
// -------------------------------------------
package helper

import (
	"sync"
	"time"
)

type Cache struct {
	data      map[string]interface{}
	expiresAt map[string]time.Time
	mu        sync.RWMutex
	ttl       time.Duration
}

func NewCache(defaultTTL time.Duration) *Cache {
	if defaultTTL <= 0 {
		defaultTTL = 30 * time.Second
	}
	return &Cache{
		data:      make(map[string]interface{}),
		expiresAt: make(map[string]time.Time),
		ttl:       defaultTTL,
	}
}

func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	value, exists := c.data[key]
	expireTime, hasExpire := c.expiresAt[key]
	c.mu.RUnlock()

	if !exists {
		return nil, false
	}

	if !hasExpire || time.Now().After(expireTime) {

		c.mu.Lock()

		if et, ok := c.expiresAt[key]; ok && (time.Now().After(et) || !hasExpire) {
			delete(c.data, key)
			delete(c.expiresAt, key)
		}
		c.mu.Unlock()
		return nil, false
	}

	return value, true
}

func (c *Cache) Set(key string, value interface{}) {
	c.SetWithTTL(key, value, c.ttl)
}

func (c *Cache) SetWithTTL(key string, value interface{}, ttl time.Duration) {
	if ttl <= 0 {
		ttl = c.ttl
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[key] = value
	c.expiresAt[key] = time.Now().Add(ttl)
}

func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.data, key)
	delete(c.expiresAt, key)
}

func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = make(map[string]interface{})
	c.expiresAt = make(map[string]time.Time)
}

func (c *Cache) CleanExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, expireTime := range c.expiresAt {
		if now.After(expireTime) {
			delete(c.data, key)
			delete(c.expiresAt, key)
		}
	}
}

func (c *Cache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.data)
}

func (c *Cache) GetStringSlice(key string) ([]string, bool) {
	value, ok := c.Get(key)
	if !ok {
		return nil, false
	}

	if slice, ok := value.([]string); ok {
		return slice, true
	}
	return nil, false
}

func (c *Cache) SetStringSlice(key string, value []string) {
	c.Set(key, value)
}

func (c *Cache) SetStringSliceWithTTL(key string, value []string, ttl time.Duration) {
	c.SetWithTTL(key, value, ttl)
}
