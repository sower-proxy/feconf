package ws

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sower-proxy/conf/reader"
)

const (
	// SchemeWS represents WebSocket URI scheme
	SchemeWS reader.Scheme = "ws"
	// SchemeWSS represents WebSocket Secure URI scheme
	SchemeWSS reader.Scheme = "wss"

	// DefaultTimeout for WebSocket connection
	DefaultTimeout = 30 * time.Second
	// DefaultRetryAttempts for failed connections
	DefaultRetryAttempts = 3
	// DefaultRetryDelay between retry attempts
	DefaultRetryDelay = 1 * time.Second
	// DefaultPingInterval for WebSocket ping
	DefaultPingInterval = 30 * time.Second
	// DefaultPongWait for WebSocket pong response
	DefaultPongWait = 60 * time.Second
	// DefaultWriteWait for WebSocket write operations
	DefaultWriteWait = 10 * time.Second
)

// init registers WebSocket readers
func init() {
	_ = reader.RegisterReader(SchemeWS, func(uri string) (reader.ConfReader, error) {
		return NewWSReader(uri)
	})
	_ = reader.RegisterReader(SchemeWSS, func(uri string) (reader.ConfReader, error) {
		return NewWSReader(uri)
	})
}

// WSConfig holds WebSocket client configuration
type WSConfig struct {
	Timeout       time.Duration
	RetryAttempts int
	RetryDelay    time.Duration
	PingInterval  time.Duration
	PongWait      time.Duration
	WriteWait     time.Duration
	Headers       map[string]string
	TLSConfig     *tls.Config
}

// WSReader implements ConfReader for WebSocket-based configuration
type WSReader struct {
	uri       string
	parsedURL *url.URL
	config    *WSConfig
	dialer    *websocket.Dialer
	mu        sync.RWMutex
	closed    bool
}

// NewWSReader creates a new WebSocket reader
func NewWSReader(uri string) (*WSReader, error) {
	u, err := reader.ParseURI(uri)
	if err != nil {
		return nil, err
	}

	if u.Scheme != string(SchemeWS) && u.Scheme != string(SchemeWSS) {
		return nil, fmt.Errorf("unsupported scheme: %s, expected: %s or %s", u.Scheme, SchemeWS, SchemeWSS)
	}

	config := &WSConfig{
		Timeout:       DefaultTimeout,
		RetryAttempts: DefaultRetryAttempts,
		RetryDelay:    DefaultRetryDelay,
		PingInterval:  DefaultPingInterval,
		PongWait:      DefaultPongWait,
		WriteWait:     DefaultWriteWait,
		Headers:       make(map[string]string),
	}

	// Parse query parameters for configuration
	if err := parseQueryConfig(u, config); err != nil {
		return nil, fmt.Errorf("failed to parse query configuration: %w", err)
	}

	dialer := &websocket.Dialer{
		HandshakeTimeout: config.Timeout,
		TLSClientConfig:  config.TLSConfig,
	}

	return &WSReader{
		uri:       uri,
		parsedURL: u,
		config:    config,
		dialer:    dialer,
	}, nil
}

// Read reads configuration data from WebSocket endpoint
func (w *WSReader) Read(ctx context.Context) ([]byte, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	if w.closed {
		return nil, fmt.Errorf("reader is closed")
	}

	return w.readWithRetry(ctx)
}

// Subscribe subscribes to WebSocket for real-time updates
func (w *WSReader) Subscribe(ctx context.Context) (<-chan *reader.ReadEvent, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.closed {
		return nil, fmt.Errorf("reader is closed")
	}

	eventChan := make(chan *reader.ReadEvent, 1)

	go w.subscribeWS(ctx, eventChan)

	return eventChan, nil
}

// Close closes the reader and cleans up resources
func (w *WSReader) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.closed {
		return nil
	}

	w.closed = true
	return nil
}

// readWithRetry performs WebSocket connection with retry mechanism for one-time read
func (w *WSReader) readWithRetry(ctx context.Context) ([]byte, error) {
	var lastErr error

	for attempt := range w.config.RetryAttempts {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(w.config.RetryDelay):
			}
		}

		data, err := w.readOnce(ctx)
		if err == nil {
			return data, nil
		}

		lastErr = err
	}

	return nil, fmt.Errorf("failed after %d attempts: %w", w.config.RetryAttempts, lastErr)
}

// readOnce performs single WebSocket connection and reads one message
func (w *WSReader) readOnce(ctx context.Context) ([]byte, error) {
	// Prepare headers
	headers := make(http.Header)
	for key, value := range w.config.Headers {
		headers.Set(key, value)
	}

	// Establish WebSocket connection
	conn, _, err := w.dialer.DialContext(ctx, w.parsedURL.String(), headers)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to WebSocket: %w", err)
	}
	defer conn.Close()

	// Set read deadline
	_ = conn.SetReadDeadline(time.Now().Add(w.config.Timeout))

	// Read one message
	_, data, err := conn.ReadMessage()
	if err != nil {
		return nil, fmt.Errorf("failed to read WebSocket message: %w", err)
	}

	return data, nil
}

// subscribeWS handles WebSocket subscription for real-time updates
func (w *WSReader) subscribeWS(ctx context.Context, eventChan chan<- *reader.ReadEvent) {
	defer close(eventChan)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		w.handleWSConnection(ctx, eventChan)

		// Wait before reconnection
		select {
		case <-ctx.Done():
			return
		case <-time.After(w.config.RetryDelay):
		}
	}
}

// handleWSConnection handles single WebSocket connection lifecycle
func (w *WSReader) handleWSConnection(ctx context.Context, eventChan chan<- *reader.ReadEvent) {
	// Prepare headers
	headers := make(http.Header)
	for key, value := range w.config.Headers {
		headers.Set(key, value)
	}

	// Establish WebSocket connection
	conn, _, err := w.dialer.DialContext(ctx, w.parsedURL.String(), headers)
	if err != nil {
		confEvent := reader.NewReadEvent(w.uri, nil, fmt.Errorf("failed to connect to WebSocket: %w", err))
		select {
		case eventChan <- confEvent:
		case <-ctx.Done():
		}
		return
	}
	defer conn.Close()

	// Set up ping/pong handlers
	_ = conn.SetReadDeadline(time.Now().Add(w.config.PongWait))
	conn.SetPongHandler(func(string) error {
		_ = conn.SetReadDeadline(time.Now().Add(w.config.PongWait))
		return nil
	})

	// Start ping ticker
	ticker := time.NewTicker(w.config.PingInterval)
	defer ticker.Stop()

	// Handle ping in separate goroutine
	pingDone := make(chan struct{})
	go func() {
		defer close(pingDone)
		for {
			select {
			case <-ticker.C:
				_ = conn.SetWriteDeadline(time.Now().Add(w.config.WriteWait))
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					return
				}
			case <-ctx.Done():
				return
			case <-pingDone:
				return
			}
		}
	}()

	// Read messages loop
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		_, data, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				confEvent := reader.NewReadEvent(w.uri, nil, fmt.Errorf("WebSocket read error: %w", err))
				select {
				case eventChan <- confEvent:
				case <-ctx.Done():
				}
			}
			return
		}

		// Send configuration event
		confEvent := reader.NewReadEvent(w.uri, data, nil)
		select {
		case eventChan <- confEvent:
		case <-ctx.Done():
			return
		}
	}
}

// parseQueryConfig parses query parameters into WebSocket configuration
func parseQueryConfig(u *url.URL, config *WSConfig) error {
	query := u.Query()

	// Parse timeout
	if timeoutStr := query.Get("timeout"); timeoutStr != "" {
		timeout, err := time.ParseDuration(timeoutStr)
		if err != nil {
			return fmt.Errorf("invalid timeout format: %w", err)
		}
		config.Timeout = timeout
	}

	// Parse retry attempts
	if retryStr := query.Get("retry_attempts"); retryStr != "" {
		var retryAttempts int
		if _, err := fmt.Sscanf(retryStr, "%d", &retryAttempts); err != nil {
			return fmt.Errorf("invalid retry_attempts format: %w", err)
		}
		if retryAttempts < 1 {
			return fmt.Errorf("retry_attempts must be at least 1")
		}
		config.RetryAttempts = retryAttempts
	}

	// Parse retry delay
	if delayStr := query.Get("retry_delay"); delayStr != "" {
		delay, err := time.ParseDuration(delayStr)
		if err != nil {
			return fmt.Errorf("invalid retry_delay format: %w", err)
		}
		config.RetryDelay = delay
	}

	// Parse ping interval
	if pingStr := query.Get("ping_interval"); pingStr != "" {
		interval, err := time.ParseDuration(pingStr)
		if err != nil {
			return fmt.Errorf("invalid ping_interval format: %w", err)
		}
		config.PingInterval = interval
	}

	// Parse pong wait
	if pongStr := query.Get("pong_wait"); pongStr != "" {
		wait, err := time.ParseDuration(pongStr)
		if err != nil {
			return fmt.Errorf("invalid pong_wait format: %w", err)
		}
		config.PongWait = wait
	}

	// Parse write wait
	if writeStr := query.Get("write_wait"); writeStr != "" {
		wait, err := time.ParseDuration(writeStr)
		if err != nil {
			return fmt.Errorf("invalid write_wait format: %w", err)
		}
		config.WriteWait = wait
	}

	// Parse headers (format: header_name=value)
	for key, values := range query {
		if strings.HasPrefix(key, "header_") {
			headerName := strings.TrimPrefix(key, "header_")
			if len(values) > 0 {
				config.Headers[headerName] = values[0]
			}
		}
	}

	// Parse TLS configuration
	if insecureStr := query.Get("tls_insecure"); insecureStr == "true" {
		if config.TLSConfig == nil {
			config.TLSConfig = &tls.Config{}
		}
		config.TLSConfig.InsecureSkipVerify = true
	}

	return nil
}
