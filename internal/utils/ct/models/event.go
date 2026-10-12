// ct/models-----------------------------------
// @file      : event.go
// @time      : 2025/01/24 21:00
// -------------------------------------------

// ============================================

package models

import (
	"fmt"
	"time"
)

type EventType string

const (
	EventTypeSubdomainDiscovered EventType = "subdomain_discovered"

	EventTypeCertificateProcessed EventType = "certificate_processed"

	EventTypeErrorOccurred EventType = "error_occurred"

	EventTypeWatcherStatusChanged EventType = "watcher_status_changed"

	EventTypeSDKStarted EventType = "sdk_started"
	EventTypeSDKStopped EventType = "sdk_stopped"
)

type Event interface {
	GetType() EventType
	GetTimestamp() time.Time
	GetID() string
}

type BaseEvent struct {
	Type      EventType `json:"type"`
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
}

func (e *BaseEvent) GetType() EventType {
	return e.Type
}

func (e *BaseEvent) GetTimestamp() time.Time {
	return e.Timestamp
}

func (e *BaseEvent) GetID() string {
	return e.ID
}

type SubdomainDiscoveredEvent struct {
	BaseEvent
	SubdomainEvent SubdomainEvent `json:"subdomain_event"`
}

type SubdomainEvent struct {
	Subdomain    string    `json:"subdomain"`
	RootDomain   string    `json:"root_domain"`
	Certificate  *CertInfo `json:"certificate"`
	DiscoveredAt time.Time `json:"discovered_at"`
	EventID      string    `json:"event_id"`
	IsNew        bool      `json:"is_new"`
}

type CertInfo struct {
	LogURL     string    `json:"log_url"`
	LogID      string    `json:"log_id"`
	Index      int64     `json:"index"`
	Timestamp  time.Time `json:"timestamp"`
	CommonName string    `json:"common_name"`
	AllDomains []string  `json:"all_domains"`
}

type CertificateProcessedEvent struct {
	BaseEvent
	Certificate    *CertInfo     `json:"certificate"`
	ProcessingTime time.Duration `json:"processing_time"`
	Success        bool          `json:"success"`
	Error          string        `json:"error,omitempty"`
}

type ErrorOccurredEvent struct {
	BaseEvent
	ErrorType    string                 `json:"error_type"`
	ErrorMessage string                 `json:"error_message"`
	Context      map[string]interface{} `json:"context"`
	Severity     string                 `json:"severity"`
	Recoverable  bool                   `json:"recoverable"`
}

type WatcherStatusChangedEvent struct {
	BaseEvent
	LogURL    string `json:"log_url"`
	OldStatus string `json:"old_status"`
	NewStatus string `json:"new_status"`
	Reason    string `json:"reason"`
}

type SDKLifecycleEvent struct {
	BaseEvent
	Action   string         `json:"action"`
	Reason   string         `json:"reason"`
	Duration *time.Duration `json:"duration,omitempty"`
	Stats    *SDKStats      `json:"stats,omitempty"`
}

type SDKStats struct {
	TotalProcessedCertificates int64         `json:"total_processed_certificates"`
	TotalDiscoveredSubdomains  int64         `json:"total_discovered_subdomains"`
	TotalErrors                int64         `json:"total_errors"`
	Uptime                     time.Duration `json:"uptime"`
}

type SubdomainCallback func(event SubdomainEvent)

type CertificateCallback func(cert *CertInfo, success bool, err error)

type ErrorCallback func(err error, context map[string]interface{})

type StatusCallback func(logURL string, oldStatus, newStatus string)

type EventHandler interface {
	HandleEvent(event Event) error
}

type EventBus struct {
	handlers map[EventType][]EventHandler
}

func NewEventBus() *EventBus {
	return &EventBus{
		handlers: make(map[EventType][]EventHandler),
	}
}

func (eb *EventBus) Subscribe(eventType EventType, handler EventHandler) {
	eb.handlers[eventType] = append(eb.handlers[eventType], handler)
}

func (eb *EventBus) Publish(event Event) error {
	handlers, exists := eb.handlers[event.GetType()]
	if !exists {
		return nil
	}

	for _, handler := range handlers {
		if err := handler.HandleEvent(event); err != nil {

			continue
		}
	}
	return nil
}

func GenerateEventID() string {
	return fmt.Sprintf("evt_%d", time.Now().UnixNano())
}
