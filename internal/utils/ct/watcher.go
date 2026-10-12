// ct-----------------------------------------
// @file      : watcher.go
// @time      : 2025/01/24 21:00
// -------------------------------------------

// ============================================

package ct

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"recon/internal/global"
	"recon/internal/utils/ct/models"
	"recon/internal/utils/ct/types"

	ct "github.com/google/certificate-transparency-go"
	"github.com/google/certificate-transparency-go/client"
	"github.com/google/certificate-transparency-go/jsonclient"
	"github.com/google/certificate-transparency-go/scanner"
	"go.uber.org/zap"
)

type CTWatcher struct {
	config      *types.LogServerConfig
	performance *types.PerformanceConfig

	streamProcessor *StreamProcessor
	checkpointSaver *AutoCheckpointSaver

	httpClient *http.Client
	ctClient   *client.LogClient
	scanner    *scanner.Scanner

	status     *models.WatcherStatus
	checkpoint *models.Checkpoint
	isRunning  int32
	lastIndex  int64

	logger *zap.Logger

	ctx    context.Context
	cancel context.CancelFunc

	stats struct {
		totalProcessed int64
		totalErrors    int64
		startTime      time.Time
	}
}

func NewCTWatcher(
	config *types.LogServerConfig,
	performance *types.PerformanceConfig,
	streamProcessor *StreamProcessor,
	checkpointSaver *AutoCheckpointSaver,
) *CTWatcher {

	ctx, cancel := context.WithCancel(context.Background())

	watcher := &CTWatcher{
		config:          config,
		performance:     performance,
		streamProcessor: streamProcessor,
		checkpointSaver: checkpointSaver,
		status: models.NewWatcherStatus(&models.LogServerConfig{
			URL:         config.URL,
			Description: config.URL,
			Operator:    "Unknown",
		}),
		logger: global.Log,
		ctx:    ctx,
		cancel: cancel,
	}

	watcher.initHTTPClient()

	return watcher
}

func (w *CTWatcher) Start() error {
	if !atomic.CompareAndSwapInt32(&w.isRunning, 0, 1) {
		return fmt.Errorf("Watcher is already running")
	}

	w.stats.startTime = time.Now()
	w.status.UpdateStatus("connecting")

	if err := w.initCTClient(); err != nil {
		atomic.StoreInt32(&w.isRunning, 0)
		w.status.UpdateStatus("error")
		return fmt.Errorf("Failed to initialize CT client: %w", err)
	}

	if err := w.loadCheckpoint(); err != nil {
		atomic.StoreInt32(&w.isRunning, 0)
		w.status.UpdateStatus("error")
		return fmt.Errorf("Failed to load checkpoint: %w", err)
	}

	if err := w.createScanner(); err != nil {
		atomic.StoreInt32(&w.isRunning, 0)
		w.status.UpdateStatus("error")
		return fmt.Errorf("Failed to create scanner: %w", err)
	}

	w.status.UpdateStatus("connected")

	go w.watchLoop()

	return nil
}

func (w *CTWatcher) Stop() error {
	if !atomic.CompareAndSwapInt32(&w.isRunning, 1, 0) {
		return nil
	}

	w.status.UpdateStatus("disconnected")

	w.cancel()

	w.saveCheckpoint()

	return nil
}

func (w *CTWatcher) GetStatus() *models.WatcherStatus {

	w.status.ProcessedCount = atomic.LoadInt64(&w.stats.totalProcessed)
	w.status.LastProcessedIndex = atomic.LoadInt64(&w.lastIndex)
	w.status.LastActivity = time.Now()

	return w.status
}

func (w *CTWatcher) watchLoop() {
	defer func() {
		atomic.StoreInt32(&w.isRunning, 0)
		w.status.UpdateStatus("stopped")
	}()

	for {
		select {
		case <-w.ctx.Done():
			return
		default:
		}

		if err := w.performScan(); err != nil {
			w.status.RecordError()
			atomic.AddInt64(&w.stats.totalErrors, 1)

			select {
			case <-w.ctx.Done():
				return
			case <-time.After(w.calculateRetryDelay()):
				continue
			}
		}

		select {
		case <-w.ctx.Done():
			return
		case <-time.After(30 * time.Second):
			continue
		}
	}
}

func (w *CTWatcher) performScan() error {
	if w.scanner == nil {
		return fmt.Errorf("Scanner is not initialized")
	}

	w.status.UpdateStatus("scanning")

	scanCtx, cancel := context.WithTimeout(w.ctx, 5*time.Minute)
	defer cancel()

	err := w.scanner.Scan(scanCtx, w.foundCertCallback, w.foundPrecertCallback)
	if err != nil {
		w.status.UpdateStatus("error")
		return fmt.Errorf("Scan failed: %w", err)
	}

	w.status.UpdateStatus("connected")
	return nil
}

func (w *CTWatcher) foundCertCallback(rawEntry *ct.RawLogEntry) {
	w.processRawEntry(rawEntry, "X509LogEntry")
}

func (w *CTWatcher) foundPrecertCallback(rawEntry *ct.RawLogEntry) {
	w.processRawEntry(rawEntry, "PrecertLogEntry")
}

func (w *CTWatcher) processRawEntry(rawEntry *ct.RawLogEntry, entryType string) {

	timestamp := time.Unix(int64(rawEntry.Leaf.TimestampedEntry.Timestamp)/1000, 0)

	for {
		currentIndex := atomic.LoadInt64(&w.lastIndex)
		if rawEntry.Index <= currentIndex {

			break
		}
		if atomic.CompareAndSwapInt64(&w.lastIndex, currentIndex, rawEntry.Index) {

			break
		}

	}

	var certData []byte
	switch entryType {
	case "X509LogEntry":
		if rawEntry.Leaf.TimestampedEntry.X509Entry == nil {
			w.logger.Warn("X509Entry is empty; skipping", zap.Int64("index", rawEntry.Index))
			atomic.AddInt64(&w.stats.totalProcessed, 1)
			w.status.RecordProcessing(1)
			w.updateCheckpoint(rawEntry.Index)
			return
		}
		certData = rawEntry.Leaf.TimestampedEntry.X509Entry.Data
	case "PrecertLogEntry":

		w.logger.Debug("Skipping precertificate", zap.Int64("index", rawEntry.Index))
		atomic.AddInt64(&w.stats.totalProcessed, 1)
		w.status.RecordProcessing(1)
		w.updateCheckpoint(rawEntry.Index)
		return
	default:
		w.logger.Warn("Unknown certificate type", zap.String("type", entryType), zap.Int64("index", rawEntry.Index))
		atomic.AddInt64(&w.stats.totalProcessed, 1)
		w.status.RecordProcessing(1)
		w.updateCheckpoint(rawEntry.Index)
		return
	}

	err := w.streamProcessor.ProcessCertificate(
		certData,
		rawEntry.Index,
		timestamp,
		w.config.URL,
	)

	atomic.AddInt64(&w.stats.totalProcessed, 1)
	w.status.RecordProcessing(1)

	if err != nil {
		atomic.AddInt64(&w.stats.totalErrors, 1)
		w.status.RecordError()

		w.updateCheckpoint(rawEntry.Index)
		return
	}

	w.updateCheckpoint(rawEntry.Index)
}

func (w *CTWatcher) updateCheckpoint(index int64) {
	if w.checkpoint != nil {
		w.checkpoint.Update(index, 0)
		w.checkpointSaver.UpdateCheckpoint(w.checkpoint)
	}
}

func (w *CTWatcher) saveCheckpoint() {
	if w.checkpointSaver != nil && w.checkpoint != nil {
		w.checkpointSaver.UpdateCheckpoint(w.checkpoint)
	}
}

func (w *CTWatcher) loadCheckpoint() error {
	if w.checkpointSaver == nil {
		return fmt.Errorf("Checkpoint saver is not initialized")
	}

	cp, err := w.checkpointSaver.manager.LoadCheckpoint(w.config.URL)
	if err != nil {
		return err
	}

	w.checkpoint = cp

	if cp.LastProcessedIndex == 0 {

		if w.ctClient != nil {
			sth, err := w.ctClient.GetSTH(context.Background())
			if err != nil {
				w.logger.Warn("Failed to fetch STH; continuing from the default index", zap.Error(err))

				atomic.StoreInt64(&w.lastIndex, 0)
			} else {

				treeSize := int64(sth.TreeSize)
				atomic.StoreInt64(&w.lastIndex, treeSize)
				w.logger.Info("First run: starting from current tree size",
					zap.Int64("tree_size", treeSize),
					zap.String("url", w.config.URL))
			}
		} else {

			atomic.StoreInt64(&w.lastIndex, 0)
			w.logger.Info("First run: starting from the default index",
				zap.String("url", w.config.URL))
		}
	} else {

		atomic.StoreInt64(&w.lastIndex, cp.LastProcessedIndex)
		w.logger.Info("Resuming from the saved checkpoint",
			zap.Int64("last_index", cp.LastProcessedIndex),
			zap.String("url", w.config.URL))
	}

	return nil
}

func (w *CTWatcher) initHTTPClient() {
	w.httpClient = &http.Client{
		Timeout: w.config.Timeout,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: false,
			},
			MaxIdleConns:        10,
			MaxIdleConnsPerHost: 2,
			IdleConnTimeout:     90 * time.Second,
		},
	}
}

func (w *CTWatcher) initCTClient() error {
	userAgent := "CT-Watcher/1.0"

	jsonClient, err := client.New(w.config.URL, w.httpClient, jsonclient.Options{
		UserAgent: userAgent,
	})
	if err != nil {
		return fmt.Errorf("Failed to create JSON client: %w", err)
	}

	w.ctClient = jsonClient
	return nil
}

func (w *CTWatcher) createScanner() error {
	if w.ctClient == nil {
		return fmt.Errorf("CT client is not initialized")
	}

	scannerOptions := &scanner.ScannerOptions{
		FetcherOptions: scanner.FetcherOptions{
			BatchSize:     w.config.BatchSize,
			ParallelFetch: w.config.ParallelFetch,
			StartIndex:    w.lastIndex,
			Continuous:    true,
		},
		Matcher:     scanner.MatchAll{},
		PrecertOnly: false,
		NumWorkers:  1,
		BufferSize:  w.config.BufferSize,
	}

	w.scanner = scanner.NewScanner(w.ctClient, *scannerOptions)

	return nil
}

func (w *CTWatcher) calculateRetryDelay() time.Duration {

	retryCount := w.status.ErrorCount
	if retryCount > 10 {
		retryCount = 10
	}

	delay := time.Duration(retryCount) * 5 * time.Second
	if delay > 5*time.Minute {
		delay = 5 * time.Minute
	}

	return delay
}

func (w *CTWatcher) GetStats() *WatcherStats {
	return &WatcherStats{
		TotalProcessed: atomic.LoadInt64(&w.stats.totalProcessed),
		TotalErrors:    atomic.LoadInt64(&w.stats.totalErrors),
		Uptime:         time.Since(w.stats.startTime),
		CurrentIndex:   atomic.LoadInt64(&w.lastIndex),
	}
}

type WatcherStats struct {
	TotalProcessed int64         `json:"total_processed"`
	TotalErrors    int64         `json:"total_errors"`
	Uptime         time.Duration `json:"uptime"`
	CurrentIndex   int64         `json:"current_index"`
}
