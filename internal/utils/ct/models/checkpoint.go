// ct/models-----------------------------------
// @file      : checkpoint.go
// @time      : 2025/01/24 21:00
// -------------------------------------------

// ============================================

package models

import (
	"time"
)

type Checkpoint struct {
	URL   string `json:"url" redis:"url"`
	LogID string `json:"log_id" redis:"log_id"`

	LastProcessedIndex int64 `json:"last_processed_index" redis:"last_processed_index"`
	STHTreeSize        int64 `json:"sth_tree_size" redis:"sth_tree_size"`

	Timestamp  time.Time `json:"timestamp" redis:"timestamp"`
	Version    int       `json:"version" redis:"version"`
	SDKVersion string    `json:"sdk_version" redis:"sdk_version"`
}

func NewCheckpoint(url string) *Checkpoint {
	return &Checkpoint{
		URL:       url,
		Version:   1,
		Timestamp: time.Now(),
	}
}

func (cp *Checkpoint) Update(lastIndex, sthSize int64) {
	cp.LastProcessedIndex = lastIndex
	cp.STHTreeSize = sthSize
	cp.Timestamp = time.Now()
}

func (cp *Checkpoint) IsValid() bool {
	return cp.URL != "" && cp.LastProcessedIndex >= 0
}

type CheckpointManager interface {
	SaveCheckpoint(cp *Checkpoint) error

	LoadCheckpoint(url string) (*Checkpoint, error)

	DeleteCheckpoint(url string) error

	ListCheckpoints() ([]*Checkpoint, error)

	CleanupExpired(olderThan time.Duration) error
}

type WatcherStatus struct {
	LogServer *LogServerConfig `json:"log_server"`
	LogID     string           `json:"log_id"`

	IsConnected bool   `json:"is_connected"`
	Status      string `json:"status"`

	LastProcessedIndex int64 `json:"last_processed_index"`
	ProcessedCount     int64 `json:"processed_count"`
	ErrorCount         int64 `json:"error_count"`
	RetryCount         int64 `json:"retry_count"`

	StartedAt     time.Time  `json:"started_at"`
	LastActivity  time.Time  `json:"last_activity"`
	LastErrorTime *time.Time `json:"last_error_time"`
	NextRetryTime *time.Time `json:"next_retry_time"`

	ProcessingRate float64       `json:"processing_rate"`
	AverageLatency time.Duration `json:"average_latency"`
}

func NewWatcherStatus(logServer *LogServerConfig) *WatcherStatus {
	now := time.Now()
	return &WatcherStatus{
		LogServer:    logServer,
		Status:       "disconnected",
		StartedAt:    now,
		LastActivity: now,
	}
}

func (ws *WatcherStatus) UpdateStatus(status string) {
	ws.Status = status
	ws.LastActivity = time.Now()

	if status == "connected" {
		ws.IsConnected = true
	} else if status == "disconnected" || status == "error" {
		ws.IsConnected = false
	}
}

func (ws *WatcherStatus) RecordProcessing(count int64) {
	ws.ProcessedCount += count
	ws.LastActivity = time.Now()
}

func (ws *WatcherStatus) RecordError() {
	now := time.Now()
	ws.ErrorCount++
	ws.LastErrorTime = &now
	ws.LastActivity = now
}

type SDKStatus struct {
	IsRunning bool          `json:"is_running"`
	StartTime time.Time     `json:"start_time"`
	Uptime    time.Duration `json:"uptime"`

	TotalProcessedCertificates int64 `json:"total_processed_certificates"`
	TotalDiscoveredSubdomains  int64 `json:"total_discovered_subdomains"`
	TotalErrors                int64 `json:"total_errors"`

	MemoryUsage    int64   `json:"memory_usage"`
	CPUUsage       float64 `json:"cpu_usage"`
	GoroutineCount int     `json:"goroutine_count"`

	ActiveWatchers  int                       `json:"active_watchers"`
	WatcherStatuses map[string]*WatcherStatus `json:"watcher_statuses"`

	TotalDomains   int64   `json:"total_domains"`
	MatchedDomains int64   `json:"matched_domains"`
	MatchRate      float64 `json:"match_rate"`
}

func NewSDKStatus() *SDKStatus {
	return &SDKStatus{
		StartTime:       time.Now(),
		WatcherStatuses: make(map[string]*WatcherStatus),
	}
}

func (ss *SDKStatus) UpdateStats() {
	ss.Uptime = time.Since(ss.StartTime)

	activeCount := 0
	for _, status := range ss.WatcherStatuses {
		if status.IsConnected {
			activeCount++
		}
	}
	ss.ActiveWatchers = activeCount

	if ss.TotalProcessedCertificates > 0 {
		ss.MatchRate = float64(ss.TotalDiscoveredSubdomains) / float64(ss.TotalProcessedCertificates)
	}
}

type LogStatus struct {
	LogServer      *LogServerConfig `json:"log_server"`
	IsConnected    bool             `json:"is_connected"`
	LastIndex      int64            `json:"last_index"`
	ErrorCount     int64            `json:"error_count"`
	ProcessedCount int64            `json:"processed_count"`
	Status         string           `json:"status"`
}

type LogServerConfig struct {
	URL         string `json:"url"`
	Description string `json:"description"`
	Operator    string `json:"operator"`
}

type ProcessingStats struct {
	TimeWindow time.Duration `json:"time_window"`

	CertificatesPerSecond float64       `json:"certificates_per_second"`
	SubdomainsPerSecond   float64       `json:"subdomains_per_second"`
	ErrorRate             float64       `json:"error_rate"`
	AverageLatency        time.Duration `json:"average_latency"`

	MemoryUsage int64   `json:"memory_usage"`
	CPUUsage    float64 `json:"cpu_usage"`
	GCStats     GCStats `json:"gc_stats"`
}

type GCStats struct {
	NumGC         uint32    `json:"num_gc"`
	LastGCTime    time.Time `json:"last_gc_time"`
	GCCPUFraction float64   `json:"gc_cpu_fraction"`
}
