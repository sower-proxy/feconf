package reader

import (
	"context"
	"time"
)

// ConfReader defines configuration reader interface
type ConfReader interface {
	// Read reads configuration data
	Read(ctx context.Context) ([]byte, error)

	// Subscribe subscribes to configuration changes and returns update channel
	Subscribe(ctx context.Context) (<-chan *ReadEvent, error)

	// Close closes the reader and cleans up resources
	Close() error
}

// ReadEvent represents configuration update event
type ReadEvent struct {
	SourceURI string    `json:"source_uri"`
	Timestamp time.Time `json:"timestamp"`
	Data      []byte    `json:"data"`
	Error     error     `json:"error,omitempty"`
}

// NewReadEvent creates a new configuration event with validation
func NewReadEvent(sourceURI string, data []byte, err error) *ReadEvent {
	event := &ReadEvent{
		SourceURI: sourceURI,
		Timestamp: time.Now(),
		Data:      data,
		Error:     err,
	}

	return event
}

// IsValid checks if the configuration event is valid
func (e *ReadEvent) IsValid() bool {
	return e != nil && e.Error == nil && len(e.Data) > 0
}
