package http

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/sower-proxy/feconf/reader"
)

const (
	// SchemeHTTP represents HTTP URI scheme
	SchemeHTTP reader.Scheme = "http"
	// SchemeHTTPS represents HTTPS URI scheme
	SchemeHTTPS reader.Scheme = "https"

	// DefaultTimeout for HTTP requests
	DefaultTimeout = 30 * time.Second
	// DefaultRetryAttempts for failed requests
	DefaultRetryAttempts = 3
	// DefaultRetryDelay between retry attempts
	DefaultRetryDelay = 1 * time.Second
)

// init registers HTTP readers
func init() {
	_ = reader.RegisterReader(SchemeHTTP, func(uri string) (reader.ConfReader, error) {
		return NewHTTPReader(uri)
	})
	_ = reader.RegisterReader(SchemeHTTPS, func(uri string) (reader.ConfReader, error) {
		return NewHTTPReader(uri)
	})
}

// HTTPConfig holds HTTP client configuration
type HTTPConfig struct {
	Timeout       time.Duration
	RetryAttempts int
	RetryDelay    time.Duration
	Headers       map[string]string
	TLSConfig     *tls.Config
}

// HTTPReader implements ConfReader for HTTP-based configuration
type HTTPReader struct {
	uri       string
	parsedURL *url.URL
	client    *http.Client
	config    *HTTPConfig
	mu        sync.RWMutex
	closed    bool
}

// NewHTTPReader creates a new HTTP reader
func NewHTTPReader(uri string) (*HTTPReader, error) {
	u, err := reader.ParseURI(uri)
	if err != nil {
		return nil, err
	}

	if u.Scheme != string(SchemeHTTP) && u.Scheme != string(SchemeHTTPS) {
		return nil, fmt.Errorf("unsupported scheme: %s, expected: %s or %s", u.Scheme, SchemeHTTP, SchemeHTTPS)
	}

	config := &HTTPConfig{
		Timeout:       DefaultTimeout,
		RetryAttempts: DefaultRetryAttempts,
		RetryDelay:    DefaultRetryDelay,
		Headers:       make(map[string]string),
	}

	// Parse query parameters for configuration
	if err := parseQueryConfig(u, config); err != nil {
		return nil, fmt.Errorf("failed to parse query configuration: %w", err)
	}

	transport := &http.Transport{
		TLSClientConfig: config.TLSConfig,
	}

	client := &http.Client{
		Timeout:   config.Timeout,
		Transport: transport,
	}

	return &HTTPReader{
		uri:       uri,
		parsedURL: u,
		client:    client,
		config:    config,
	}, nil
}

// Read reads configuration data from HTTP endpoint
func (h *HTTPReader) Read(ctx context.Context) ([]byte, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if h.closed {
		return nil, fmt.Errorf("reader is closed")
	}

	return h.fetchWithRetry(ctx)
}

// Subscribe subscribes to HTTP endpoint with SSE support for real-time updates
func (h *HTTPReader) Subscribe(ctx context.Context) (<-chan *reader.ReadEvent, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.closed {
		return nil, fmt.Errorf("reader is closed")
	}

	eventChan := make(chan *reader.ReadEvent, 1)

	go h.subscribeSSE(ctx, eventChan)

	return eventChan, nil
}

// Close closes the reader and cleans up resources
func (h *HTTPReader) Close() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.closed {
		return nil
	}

	h.closed = true
	h.client.CloseIdleConnections()

	return nil
}

// fetchWithRetry performs HTTP GET with retry mechanism
func (h *HTTPReader) fetchWithRetry(ctx context.Context) ([]byte, error) {
	var lastErr error

	for attempt := 0; attempt < h.config.RetryAttempts; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(h.config.RetryDelay):
			}
		}

		data, err := h.fetch(ctx)
		if err == nil {
			return data, nil
		}

		lastErr = err
	}

	return nil, fmt.Errorf("failed after %d attempts: %w", h.config.RetryAttempts, lastErr)
}

// fetch performs single HTTP GET request
func (h *HTTPReader) fetch(ctx context.Context) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, h.parsedURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set custom headers
	for key, value := range h.config.Headers {
		req.Header.Set(key, value)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP request failed with status: %d %s", resp.StatusCode, resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return data, nil
}

// subscribeSSE handles Server-Sent Events subscription
func (h *HTTPReader) subscribeSSE(ctx context.Context, eventChan chan<- *reader.ReadEvent) {
	defer close(eventChan)

	// Create SSE request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, h.parsedURL.String(), nil)
	if err != nil {
		confEvent := reader.NewReadEvent(h.uri, nil, fmt.Errorf("failed to create SSE request: %w", err))
		select {
		case eventChan <- confEvent:
		case <-ctx.Done():
		}
		return
	}

	// Set SSE headers
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")

	// Set custom headers
	for key, value := range h.config.Headers {
		req.Header.Set(key, value)
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		resp, err := h.client.Do(req)
		if err != nil {
			confEvent := reader.NewReadEvent(h.uri, nil, fmt.Errorf("SSE connection failed: %w", err))
			select {
			case eventChan <- confEvent:
			case <-ctx.Done():
				return
			}

			// Wait before retry
			select {
			case <-ctx.Done():
				return
			case <-time.After(h.config.RetryDelay):
				continue
			}
		}

		if resp.StatusCode != http.StatusOK {
			_ = resp.Body.Close()
			confEvent := reader.NewReadEvent(h.uri, nil, fmt.Errorf("SSE request failed with status: %d %s", resp.StatusCode, resp.Status))
			select {
			case eventChan <- confEvent:
			case <-ctx.Done():
				return
			}

			// Wait before retry
			select {
			case <-ctx.Done():
				return
			case <-time.After(h.config.RetryDelay):
				continue
			}
		}

		// Process SSE stream
		h.processSSEStream(ctx, resp, eventChan)
		_ = resp.Body.Close()

		// Wait before reconnection
		select {
		case <-ctx.Done():
			return
		case <-time.After(h.config.RetryDelay):
		}
	}
}

// processSSEStream processes Server-Sent Events stream
func (h *HTTPReader) processSSEStream(ctx context.Context, resp *http.Response, eventChan chan<- *reader.ReadEvent) {
	scanner := bufio.NewScanner(resp.Body)
	var eventData strings.Builder

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return
		default:
		}

		line := scanner.Text()

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, ":") {
			continue
		}

		// Handle data lines
		if strings.HasPrefix(line, "data:") {
			data := strings.TrimPrefix(line, "data:")
			data = strings.TrimSpace(data)
			if eventData.Len() > 0 {
				eventData.WriteString("\n")
			}
			eventData.WriteString(data)
		} else if line == "" {
			// Empty line indicates end of event
			if eventData.Len() > 0 {
				confEvent := reader.NewReadEvent(h.uri, []byte(eventData.String()), nil)
				select {
				case eventChan <- confEvent:
				case <-ctx.Done():
					return
				}
				eventData.Reset()
			}
		}
	}

	if err := scanner.Err(); err != nil {
		confEvent := reader.NewReadEvent(h.uri, nil, fmt.Errorf("SSE stream error: %w", err))
		select {
		case eventChan <- confEvent:
		case <-ctx.Done():
			return
		}
	}
}

// parseQueryConfig parses query parameters into HTTP configuration
func parseQueryConfig(u *url.URL, config *HTTPConfig) error {
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

	// Parse headers (format: header_name=value)
	for key, values := range query {
		if strings.HasPrefix(key, "header_") {
			headerName := strings.TrimPrefix(key, "header_")
			if len(values) > 0 {
				config.Headers[headerName] = values[0]
			}
		}
	}

	// Parse basic auth credentials from userinfo
	if u.User != nil {
		username := u.User.Username()
		password, _ := u.User.Password()
		if username != "" {
			basicAuth := fmt.Sprintf("%s:%s", username, password)
			config.Headers["Authorization"] = fmt.Sprintf("Basic %s",
				basicAuthEncode(basicAuth))
		}
	}

	// Parse TLS configuration
	if insecureStr := query.Get("tls_insecure"); insecureStr == "true" {
		if config.TLSConfig == nil {
			config.TLSConfig = &tls.Config{
				MinVersion: tls.VersionTLS12,
			}
		} else {
			// Ensure minimum TLS version is set
			if config.TLSConfig.MinVersion == 0 {
				config.TLSConfig.MinVersion = tls.VersionTLS12
			}
		}
		config.TLSConfig.InsecureSkipVerify = true
	} else if u.Scheme == string(SchemeHTTPS) && config.TLSConfig == nil {
		// For HTTPS connections, ensure minimum TLS version
		config.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	} else if config.TLSConfig != nil && config.TLSConfig.MinVersion == 0 {
		// If TLS config exists but no minimum version is set, use TLS 1.2
		config.TLSConfig.MinVersion = tls.VersionTLS12
	}

	return nil
}

// basicAuthEncode encodes username:password for basic authentication
func basicAuthEncode(credentials string) string {
	return base64.StdEncoding.EncodeToString([]byte(credentials))
}
