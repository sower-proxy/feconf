package reader

import (
	"context"
	"time"
)

// ConfReader defines configuration reader interface
type ConfReader interface {
	// Read reads configuration data
	Read(ctx context.Context) (*ConfEvent, error)

	// Subscribe subscribes to configuration changes and returns update channel
	Subscribe(ctx context.Context) (<-chan *ConfEvent, error)

	// Close closes the reader and cleans up resources
	Close() error
}

// ConfEvent represents configuration update event
type ConfEvent struct {
	SourceURI string    `json:"source_uri"`
	Timestamp time.Time `json:"timestamp"`
	Data      []byte    `json:"data"`
	Error     error     `json:"error,omitempty"`
}

// NewConfEvent creates a new configuration event with validation
func NewConfEvent(sourceURI string, data []byte, err error) *ConfEvent {
	event := &ConfEvent{
		SourceURI: sourceURI,
		Timestamp: time.Now(),
		Data:      data,
		Error:     err,
	}

	return event
}

// IsValid checks if the configuration event is valid
func (e *ConfEvent) IsValid() bool {
	return e != nil && e.Error == nil && len(e.Data) > 0
}
