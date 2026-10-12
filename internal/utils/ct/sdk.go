// ct-----------------------------------------
// @file      : sdk.go
// @time      : 2025/01/24 21:00
// -------------------------------------------

// ============================================

package ct

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"recon/internal/utils/ct/models"
	"recon/internal/utils/ct/types"
)

type CTWatcherSDK struct {
	opts *types.SDKOptions

	domainMatcher     *DomainMatcher
	streamProcessor   *StreamProcessor
	checkpointSaver   *AutoCheckpointSaver
	checkpointManager *RedisCheckpointManager

	watchers map[string]*CTWatcher
	wg       sync.WaitGroup
	mu       sync.RWMutex

	isRunning int32
	status    *models.SDKStatus
	ctx       context.Context
	cancel    context.CancelFunc

	eventBus *models.EventBus

	subdomainCallbacks []models.SubdomainCallback
	errorCallbacks     []models.ErrorCallback
}

func NewCTWatcherSDK(opts *types.SDKOptions) (*CTWatcherSDK, error) {
	if opts == nil {
		opts = types.DefaultSDKOptions()
	}

	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("Configuration validation failed: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	sdk := &CTWatcherSDK{
		opts:               opts,
		watchers:           make(map[string]*CTWatcher),
		status:             models.NewSDKStatus(),
		ctx:                ctx,
		cancel:             cancel,
		eventBus:           models.NewEventBus(),
		subdomainCallbacks: make([]models.SubdomainCallback, 0),
		errorCallbacks:     make([]models.ErrorCallback, 0),
	}

	checkpointManager, err := NewRedisCheckpointManager(opts.Redis.KeyPrefix)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("Failed to initialize Redis checkpoint manager: %w", err)
	}
	sdk.checkpointManager = checkpointManager

	sdk.checkpointSaver = NewAutoCheckpointSaver(checkpointManager, 10*time.Minute)

	sdk.domainMatcher = NewDomainMatcher(opts.DomainMatcher)

	sdk.streamProcessor = NewStreamProcessor(sdk.domainMatcher, sdk.eventBus)

	for _, domain := range opts.Domains {
		if err := sdk.domainMatcher.AddDomain(domain); err != nil {

			continue
		}
	}

	sdk.eventBus.Subscribe(models.EventTypeSubdomainDiscovered, &SubdomainEventHandler{sdk: sdk})
	sdk.eventBus.Subscribe(models.EventTypeCertificateProcessed, &CertificateEventHandler{sdk: sdk})
	sdk.eventBus.Subscribe(models.EventTypeErrorOccurred, &ErrorEventHandler{sdk: sdk})

	return sdk, nil
}

func (sdk *CTWatcherSDK) Start(ctx context.Context) error {
	if !atomic.CompareAndSwapInt32(&sdk.isRunning, 0, 1) {
		return fmt.Errorf("SDK is already running")
	}

	sdk.status.StartTime = time.Now()
	sdk.status.IsRunning = true

	sdkCtx, cancel := context.WithCancel(ctx)
	sdk.cancel = cancel

	for _, logServer := range sdk.opts.LogServers {
		if err := sdk.addWatcher(logServer); err != nil {

			sdk.handleError(fmt.Errorf("Failed to create watcher [%s]: %w", logServer.URL, err), nil)
			continue
		}
	}

	if err := sdk.checkpointSaver.Start(); err != nil {
		atomic.StoreInt32(&sdk.isRunning, 0)
		return fmt.Errorf("Failed to start checkpoint saver: %w", err)
	}

	if len(sdk.watchers) == 0 {
		atomic.StoreInt32(&sdk.isRunning, 0)
		sdk.checkpointSaver.Stop()
		return fmt.Errorf("No watcher was created successfully")
	}

	go func() {
		<-sdkCtx.Done()
		sdk.stopInternal()
	}()

	return nil
}

func (sdk *CTWatcherSDK) Stop() error {
	if !atomic.CompareAndSwapInt32(&sdk.isRunning, 1, 0) {
		return nil
	}

	sdk.cancel()
	return nil
}

func (sdk *CTWatcherSDK) stopInternal() {
	sdk.mu.Lock()
	defer sdk.mu.Unlock()

	for url, watcher := range sdk.watchers {
		if err := watcher.Stop(); err != nil {
			sdk.handleError(fmt.Errorf("Failed to stop watcher [%s]: %w", url, err), nil)
		}
	}

	sdk.checkpointSaver.Stop()

	sdk.checkpointManager.Close()

	sdk.status.IsRunning = false
}

func (sdk *CTWatcherSDK) AddDomains(domains []string) error {
	for _, domain := range domains {
		if err := sdk.domainMatcher.AddDomain(domain); err != nil {
			return fmt.Errorf("Failed to add domain [%s]: %w", domain, err)
		}
	}
	return nil
}

func (sdk *CTWatcherSDK) RemoveDomains(domains []string) error {
	for _, domain := range domains {
		if err := sdk.domainMatcher.RemoveDomain(domain); err != nil {
			return fmt.Errorf("Failed to remove domain [%s]: %w", domain, err)
		}
	}
	return nil
}

func (sdk *CTWatcherSDK) GetWatchedDomains() []string {
	return sdk.domainMatcher.GetAllDomains()
}

func (sdk *CTWatcherSDK) GetStatus() *types.Status {
	sdk.mu.RLock()
	defer sdk.mu.RUnlock()

	status := &types.Status{
		IsRunning:             atomic.LoadInt32(&sdk.isRunning) == 1,
		StartTime:             sdk.status.StartTime,
		Uptime:                time.Since(sdk.status.StartTime),
		ProcessedCertificates: sdk.status.TotalProcessedCertificates,
		DiscoveredSubdomains:  sdk.status.TotalDiscoveredSubdomains,
		ActiveWatchers:        len(sdk.watchers),
		RecentErrors:          make([]string, 0),
		WatcherStatuses:       make(map[string]*models.LogStatus),
	}

	for url, watcher := range sdk.watchers {
		watcherStatus := watcher.GetStatus()
		status.WatcherStatuses[url] = &models.LogStatus{
			LogServer:      watcherStatus.LogServer,
			IsConnected:    watcherStatus.IsConnected,
			LastIndex:      watcherStatus.LastProcessedIndex,
			ErrorCount:     watcherStatus.ErrorCount,
			ProcessedCount: watcherStatus.ProcessedCount,
			Status:         watcherStatus.Status,
		}
	}

	return status
}

func (sdk *CTWatcherSDK) GetStats() *models.DomainMatcherStats {
	if sdk.domainMatcher == nil {
		return &models.DomainMatcherStats{}
	}

	return sdk.domainMatcher.GetStats()
}

func (sdk *CTWatcherSDK) OnSubdomainFound(callback models.SubdomainCallback) error {
	if callback == nil {
		return fmt.Errorf("Callback is required")
	}

	sdk.subdomainCallbacks = append(sdk.subdomainCallbacks, callback)
	return nil
}

func (sdk *CTWatcherSDK) OnError(callback models.ErrorCallback) error {
	if callback == nil {
		return fmt.Errorf("Callback is required")
	}

	sdk.errorCallbacks = append(sdk.errorCallbacks, callback)
	return nil
}

func (sdk *CTWatcherSDK) addWatcher(logServer *types.LogServerConfig) error {
	sdk.mu.Lock()
	defer sdk.mu.Unlock()

	if _, exists := sdk.watchers[logServer.URL]; exists {
		return fmt.Errorf("Watcher already exists: %s", logServer.URL)
	}

	watcher := NewCTWatcher(
		logServer,
		sdk.opts.Performance,
		sdk.streamProcessor,
		sdk.checkpointSaver,
	)

	sdk.watchers[logServer.URL] = watcher

	if err := watcher.Start(); err != nil {
		delete(sdk.watchers, logServer.URL)
		return fmt.Errorf("Failed to start watcher: %w", err)
	}

	return nil
}

type SubdomainEventHandler struct {
	sdk *CTWatcherSDK
}

func (h *SubdomainEventHandler) HandleEvent(event models.Event) error {
	if subdomainEvent, ok := event.(*models.SubdomainDiscoveredEvent); ok {
		atomic.AddInt64(&h.sdk.status.TotalDiscoveredSubdomains, 1)

		for _, callback := range h.sdk.subdomainCallbacks {
			go func(cb models.SubdomainCallback) {
				defer func() {
					if r := recover(); r != nil {
						h.sdk.handleError(fmt.Errorf("Subdomain callback failed: %v", r), nil)
					}
				}()
				cb(subdomainEvent.SubdomainEvent)
			}(callback)
		}
	}
	return nil
}

type CertificateEventHandler struct {
	sdk *CTWatcherSDK
}

func (h *CertificateEventHandler) HandleEvent(event models.Event) error {
	if certEvent, ok := event.(*models.CertificateProcessedEvent); ok {
		atomic.AddInt64(&h.sdk.status.TotalProcessedCertificates, 1)

		_ = certEvent
	}
	return nil
}

type ErrorEventHandler struct {
	sdk *CTWatcherSDK
}

func (h *ErrorEventHandler) HandleEvent(event models.Event) error {
	if errorEvent, ok := event.(*models.ErrorOccurredEvent); ok {
		atomic.AddInt64(&h.sdk.status.TotalErrors, 1)

		for _, callback := range h.sdk.errorCallbacks {
			go func(cb models.ErrorCallback) {
				defer func() {
					if r := recover(); r != nil {
						h.sdk.handleError(fmt.Errorf("Error callback failed: %v", r), nil)
					}
				}()
				cb(fmt.Errorf("%s", errorEvent.ErrorMessage), map[string]interface{}{
					"error_type": errorEvent.ErrorType,
					"severity":   errorEvent.Severity,
				})
			}(callback)
		}
	}
	return nil
}

func (sdk *CTWatcherSDK) handleError(err error, context map[string]interface{}) {
	errorEvent := &models.ErrorOccurredEvent{
		BaseEvent: models.BaseEvent{
			Type:      models.EventTypeErrorOccurred,
			ID:        models.GenerateEventID(),
			Timestamp: time.Now(),
		},
		ErrorType:    "internal_error",
		ErrorMessage: err.Error(),
		Context:      context,
		Severity:     "medium",
		Recoverable:  true,
	}

	sdk.eventBus.Publish(errorEvent)
}

func (sdk *CTWatcherSDK) IsRunning() bool {
	return atomic.LoadInt32(&sdk.isRunning) == 1
}

func (sdk *CTWatcherSDK) Close() error {
	if sdk.IsRunning() {
		sdk.Stop()
	}

	if sdk.checkpointManager != nil {
		sdk.checkpointManager.Close()
	}

	return nil
}
