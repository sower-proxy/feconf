package ws

import (
	"context"
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// TestNewWSReader tests WebSocket reader creation
func TestNewWSReader(t *testing.T) {
	tests := []struct {
		name        string
		uri         string
		expectError bool
	}{
		{
			name:        "valid ws URI",
			uri:         "ws://localhost:8080/config",
			expectError: false,
		},
		{
			name:        "valid wss URI",
			uri:         "wss://localhost:8080/config",
			expectError: false,
		},
		{
			name:        "invalid scheme",
			uri:         "http://localhost:8080/config",
			expectError: true,
		},
		{
			name:        "empty URI",
			uri:         "",
			expectError: true,
		},
		{
			name:        "URI with query parameters",
			uri:         "ws://localhost:8080/config?timeout=5s&retry_attempts=2",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader, err := NewWSReader(tt.uri)
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if reader == nil {
					t.Errorf("expected reader, got nil")
				}
				if reader != nil {
					reader.Close()
				}
			}
		})
	}
}

// TestWSReader_Read tests WebSocket reader Read functionality
func TestWSReader_Read(t *testing.T) {
	// Create test WebSocket server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Errorf("failed to upgrade connection: %v", err)
			return
		}
		defer conn.Close()

		// Send test data
		testData := `{"key": "value", "number": 42}`
		err = conn.WriteMessage(websocket.TextMessage, []byte(testData))
		if err != nil {
			t.Errorf("failed to write message: %v", err)
		}
	}))
	defer server.Close()

	// Convert HTTP test server URL to WebSocket URL
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/config"

	reader, err := NewWSReader(wsURL)
	if err != nil {
		t.Fatalf("failed to create WebSocket reader: %v", err)
	}
	defer reader.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	data, err := reader.Read(ctx)
	if err != nil {
		t.Errorf("failed to read data: %v", err)
	}

	expected := `{"key": "value", "number": 42}`
	if string(data) != expected {
		t.Errorf("expected %s, got %s", expected, string(data))
	}
}

// TestWSReader_Subscribe tests WebSocket reader Subscribe functionality
func TestWSReader_Subscribe(t *testing.T) {
	// Create test WebSocket server that sends multiple messages
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Errorf("failed to upgrade connection: %v", err)
			return
		}
		defer conn.Close()

		// Send multiple test messages
		messages := []string{
			`{"update": 1}`,
			`{"update": 2}`,
			`{"update": 3}`,
		}

		for i, msg := range messages {
			time.Sleep(100 * time.Millisecond) // Small delay between messages
			err = conn.WriteMessage(websocket.TextMessage, []byte(msg))
			if err != nil {
				t.Errorf("failed to write message %d: %v", i, err)
				return
			}
		}
	}))
	defer server.Close()

	// Convert HTTP test server URL to WebSocket URL
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/config"

	reader, err := NewWSReader(wsURL)
	if err != nil {
		t.Fatalf("failed to create WebSocket reader: %v", err)
	}
	defer reader.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	eventChan, err := reader.Subscribe(ctx)
	if err != nil {
		t.Fatalf("failed to subscribe: %v", err)
	}

	receivedCount := 0
	expectedMessages := []string{
		`{"update": 1}`,
		`{"update": 2}`,
		`{"update": 3}`,
	}

	for {
		select {
		case event := <-eventChan:
			if event == nil {
				// Channel closed
				if receivedCount != len(expectedMessages) {
					t.Errorf("expected %d messages, got %d", len(expectedMessages), receivedCount)
				}
				return
			}

			if event.Error != nil {
				t.Errorf("received error event: %v", event.Error)
				continue
			}

			if !event.IsValid() {
				t.Errorf("received invalid event: %+v", event)
				continue
			}

			if receivedCount < len(expectedMessages) {
				expected := expectedMessages[receivedCount]
				if string(event.Data) != expected {
					t.Errorf("message %d: expected %s, got %s", receivedCount, expected, string(event.Data))
				}
			}

			receivedCount++

			// Stop after receiving all expected messages
			if receivedCount >= len(expectedMessages) {
				return
			}

		case <-ctx.Done():
			t.Errorf("test timeout, received %d out of %d messages", receivedCount, len(expectedMessages))
			return
		}
	}
}

// TestWSReader_Close tests WebSocket reader Close functionality
func TestWSReader_Close(t *testing.T) {
	reader, err := NewWSReader("ws://localhost:8080/config")
	if err != nil {
		t.Fatalf("failed to create WebSocket reader: %v", err)
	}

	err = reader.Close()
	if err != nil {
		t.Errorf("failed to close reader: %v", err)
	}

	// Test that operations fail after close
	ctx := context.Background()
	_, err = reader.Read(ctx)
	if err == nil {
		t.Errorf("expected error when reading from closed reader")
	}

	_, err = reader.Subscribe(ctx)
	if err == nil {
		t.Errorf("expected error when subscribing to closed reader")
	}
}

// TestParseQueryConfig tests query parameter parsing
func TestParseQueryConfig(t *testing.T) {
	tests := []struct {
		name     string
		uri      string
		expected *WSConfig
	}{
		{
			name: "default config",
			uri:  "ws://localhost/config",
			expected: &WSConfig{
				Timeout:       DefaultTimeout,
				RetryAttempts: DefaultRetryAttempts,
				RetryDelay:    DefaultRetryDelay,
				PingInterval:  DefaultPingInterval,
				PongWait:      DefaultPongWait,
				WriteWait:     DefaultWriteWait,
				Headers:       make(map[string]string),
			},
		},
		{
			name: "custom timeout and retry",
			uri:  "ws://localhost/config?timeout=10s&retry_attempts=5&retry_delay=2s",
			expected: &WSConfig{
				Timeout:       10 * time.Second,
				RetryAttempts: 5,
				RetryDelay:    2 * time.Second,
				PingInterval:  DefaultPingInterval,
				PongWait:      DefaultPongWait,
				WriteWait:     DefaultWriteWait,
				Headers:       make(map[string]string),
			},
		},
		{
			name: "custom ping settings",
			uri:  "ws://localhost/config?ping_interval=15s&pong_wait=30s&write_wait=5s",
			expected: &WSConfig{
				Timeout:       DefaultTimeout,
				RetryAttempts: DefaultRetryAttempts,
				RetryDelay:    DefaultRetryDelay,
				PingInterval:  15 * time.Second,
				PongWait:      30 * time.Second,
				WriteWait:     5 * time.Second,
				Headers:       make(map[string]string),
			},
		},
		{
			name: "with headers and TLS",
			uri:  "ws://localhost/config?header_Authorization=Bearer%20token&header_User-Agent=test-agent&tls_insecure=true",
			expected: &WSConfig{
				Timeout:       DefaultTimeout,
				RetryAttempts: DefaultRetryAttempts,
				RetryDelay:    DefaultRetryDelay,
				PingInterval:  DefaultPingInterval,
				PongWait:      DefaultPongWait,
				WriteWait:     DefaultWriteWait,
				Headers: map[string]string{
					"Authorization": "Bearer token",
					"User-Agent":    "test-agent",
				},
				TLSConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader, err := NewWSReader(tt.uri)
			if err != nil {
				t.Fatalf("failed to create reader: %v", err)
			}
			defer reader.Close()

			config := reader.config

			// Compare basic settings
			if config.Timeout != tt.expected.Timeout {
				t.Errorf("timeout: expected %v, got %v", tt.expected.Timeout, config.Timeout)
			}
			if config.RetryAttempts != tt.expected.RetryAttempts {
				t.Errorf("retry attempts: expected %d, got %d", tt.expected.RetryAttempts, config.RetryAttempts)
			}
			if config.RetryDelay != tt.expected.RetryDelay {
				t.Errorf("retry delay: expected %v, got %v", tt.expected.RetryDelay, config.RetryDelay)
			}
			if config.PingInterval != tt.expected.PingInterval {
				t.Errorf("ping interval: expected %v, got %v", tt.expected.PingInterval, config.PingInterval)
			}
			if config.PongWait != tt.expected.PongWait {
				t.Errorf("pong wait: expected %v, got %v", tt.expected.PongWait, config.PongWait)
			}
			if config.WriteWait != tt.expected.WriteWait {
				t.Errorf("write wait: expected %v, got %v", tt.expected.WriteWait, config.WriteWait)
			}

			// Compare headers
			if len(config.Headers) != len(tt.expected.Headers) {
				t.Errorf("headers length: expected %d, got %d", len(tt.expected.Headers), len(config.Headers))
			}
			for key, expectedValue := range tt.expected.Headers {
				if actualValue, exists := config.Headers[key]; !exists || actualValue != expectedValue {
					t.Errorf("header %s: expected %s, got %s", key, expectedValue, actualValue)
				}
			}

			// Compare TLS config
			if tt.expected.TLSConfig != nil {
				if config.TLSConfig == nil {
					t.Errorf("expected TLS config, got nil")
				} else if config.TLSConfig.InsecureSkipVerify != tt.expected.TLSConfig.InsecureSkipVerify {
					t.Errorf("TLS insecure: expected %v, got %v",
						tt.expected.TLSConfig.InsecureSkipVerify,
						config.TLSConfig.InsecureSkipVerify)
				}
			} else if config.TLSConfig != nil {
				t.Errorf("expected nil TLS config, got %+v", config.TLSConfig)
			}
		})
	}
}

// TestWSReader_ErrorHandling tests error handling scenarios
func TestWSReader_ErrorHandling(t *testing.T) {
	// Test connection failure
	reader, err := NewWSReader("ws://localhost:9999/nonexistent")
	if err != nil {
		t.Fatalf("failed to create WebSocket reader: %v", err)
	}
	defer reader.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Test Read with connection failure
	_, err = reader.Read(ctx)
	if err == nil {
		t.Errorf("expected error when connecting to nonexistent server")
	}

	// Test Subscribe with connection failure
	eventChan, err := reader.Subscribe(ctx)
	if err != nil {
		t.Fatalf("failed to subscribe: %v", err)
	}

	// Should receive error event
	received := false
	for !received {
		select {
		case event := <-eventChan:
			if event == nil {
				// Channel closed, which is expected behavior
				received = true
			} else if event.Error != nil {
				// Received error event as expected
				received = true
			} else {
				// Unexpected success event, continue waiting
				continue
			}
		case <-ctx.Done():
			// Timeout is also acceptable for error handling test
			received = true
		}
	}
}
