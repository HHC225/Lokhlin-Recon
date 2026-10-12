// ct/types------------------------------------
// @file      : options.go
// @time      : 2025/01/24 21:00
// -------------------------------------------

// ============================================

package types

import (
	"fmt"
	"time"

	"recon/internal/utils/ct/models"
)

type SDKOptions struct {
	LogServers []*LogServerConfig `json:"log_servers"`

	Domains []string `json:"domains"`

	DomainMatcher *DomainMatcherConfig `json:"domain_matcher"`

	Performance *PerformanceConfig `json:"performance"`

	Redis *RedisConfig `json:"redis"`
}

type LogServerConfig struct {
	URL string `json:"url"`

	BatchSize     int `json:"batch_size"`
	ParallelFetch int `json:"parallel_fetch"`
	BufferSize    int `json:"buffer_size"`

	Timeout    time.Duration `json:"timeout"`
	MaxRetries int           `json:"max_retries"`
}

type DomainMatcherConfig struct {
	EnableBloomFilter bool `json:"enable_bloom_filter"`
	BloomFilterSize   uint `json:"bloom_filter_size"`
	BloomFilterHash   uint `json:"bloom_filter_hash"`

	EnableLRUCache bool `json:"enable_lru_cache"`
	LRUCacheSize   int  `json:"lru_cache_size"`

	CaseSensitive   bool `json:"case_sensitive"`
	MaxDomainLength int  `json:"max_domain_length"`
}

type PerformanceConfig struct {
	MaxWorkers        int `json:"max_workers"`
	ChannelBufferSize int `json:"channel_buffer_size"`

	ProcessingTimeout time.Duration `json:"processing_timeout"`
	BatchProcessSize  int           `json:"batch_process_size"`

	MaxMemoryUsage int64   `json:"max_memory_usage"`
	MaxCPUUsage    float64 `json:"max_cpu_usage"`
}

type RedisConfig struct {
	KeyPrefix string `json:"key_prefix"`
}

func DefaultSDKOptions() *SDKOptions {
	return &SDKOptions{
		LogServers: []*LogServerConfig{
			{
				URL:           "https://ct.googleapis.com/logs/argon2023/",
				BatchSize:     100,
				ParallelFetch: 1,
				BufferSize:    1000,
				Timeout:       30 * time.Second,
				MaxRetries:    3,
			},
		},
		Domains: []string{},
		DomainMatcher: &DomainMatcherConfig{
			EnableBloomFilter: true,
			BloomFilterSize:   1000000,
			BloomFilterHash:   5,
			EnableLRUCache:    true,
			LRUCacheSize:      10000,
			CaseSensitive:     false,
			MaxDomainLength:   253,
		},
		Performance: &PerformanceConfig{
			MaxWorkers:        10,
			ChannelBufferSize: 100,
			ProcessingTimeout: 5 * time.Second,
			BatchProcessSize:  50,
			MaxMemoryUsage:    512,
			MaxCPUUsage:       0.8,
		},
		Redis: &RedisConfig{
			KeyPrefix: "ct_watcher:",
		},
	}
}

func (opts *SDKOptions) Validate() error {
	if len(opts.LogServers) == 0 {
		return &ValidationError{Field: "LogServers", Message: "Configure at least one CT log server"}
	}

	for i, server := range opts.LogServers {
		if server.URL == "" {
			return &ValidationError{
				Field:   "LogServers",
				Index:   i,
				Message: "CT log URL is required",
			}
		}

		if server.BatchSize <= 0 {
			server.BatchSize = 100
		}
		if server.ParallelFetch <= 0 {
			server.ParallelFetch = 1
		}
		if server.BufferSize <= 0 {
			server.BufferSize = 1000
		}
		if server.Timeout <= 0 {
			server.Timeout = 30 * time.Second
		}
		if server.MaxRetries < 0 {
			server.MaxRetries = 3
		}
	}

	if opts.DomainMatcher == nil {
		opts.DomainMatcher = &DomainMatcherConfig{
			EnableBloomFilter: true,
			BloomFilterSize:   1000000,
			BloomFilterHash:   5,
			EnableLRUCache:    true,
			LRUCacheSize:      10000,
			CaseSensitive:     false,
			MaxDomainLength:   253,
		}
	}

	if opts.Performance == nil {
		opts.Performance = &PerformanceConfig{
			MaxWorkers:        10,
			ChannelBufferSize: 100,
			ProcessingTimeout: 5 * time.Second,
			BatchProcessSize:  50,
			MaxMemoryUsage:    512,
			MaxCPUUsage:       0.8,
		}
	}

	if opts.Redis == nil {
		opts.Redis = &RedisConfig{
			KeyPrefix: "ct_watcher:",
		}
	}

	return nil
}

type ValidationError struct {
	Field   string
	Index   int
	Message string
}

type Status struct {
	IsRunning bool          `json:"is_running"`
	StartTime time.Time     `json:"start_time"`
	Uptime    time.Duration `json:"uptime"`

	ProcessedCertificates int64 `json:"processed_certificates"`
	DiscoveredSubdomains  int64 `json:"discovered_subdomains"`
	ActiveWatchers        int   `json:"active_watchers"`

	TotalErrors  int64    `json:"total_errors"`
	RecentErrors []string `json:"recent_errors"`

	MemoryUsage    int64   `json:"memory_usage"`
	CPUUsage       float64 `json:"cpu_usage"`
	GoroutineCount int     `json:"goroutine_count"`

	WatcherStatuses map[string]*models.LogStatus `json:"watcher_statuses"`
}

func (e *ValidationError) Error() string {
	if e.Index >= 0 {
		return fmt.Sprintf("Configuration validation failed [%s[%d]]: %s", e.Field, e.Index, e.Message)
	}
	return fmt.Sprintf("Configuration validation failed [%s]: %s", e.Field, e.Message)
}
