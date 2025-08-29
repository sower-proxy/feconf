package redis

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/sower-proxy/conf/reader"
)

func TestNewRedisReader(t *testing.T) {
	tests := []struct {
		name      string
		uri       string
		wantError bool
	}{
		{
			name:      "valid redis URI with key",
			uri:       "redis://localhost:6379/config:app",
			wantError: false,
		},
		{
			name:      "valid rediss URI with key",
			uri:       "rediss://localhost:6379/config:app",
			wantError: false,
		},
		{
			name:      "URI with password",
			uri:       "redis://:password@localhost:6379/config:app",
			wantError: false,
		},
		{
			name:      "URI with database",
			uri:       "redis://localhost:6379/config:app?db=1",
			wantError: false,
		},
		{
			name:      "URI without key",
			uri:       "redis://localhost:6379/",
			wantError: true,
		},
		{
			name:      "invalid scheme",
			uri:       "http://localhost:6379/config:app",
			wantError: true,
		},
		{
			name:      "empty URI",
			uri:       "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewRedisReader(tt.uri)
			if (err != nil) != tt.wantError {
				t.Errorf("NewRedisReader() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestRedisReader_Read(t *testing.T) {
	// Setup miniredis
	s := miniredis.RunT(t)
	defer s.Close()

	// Setup test data
	testKey := "config:app"
	testValue := `{"app": "test", "version": "1.0"}`
	s.Set(testKey, testValue)

	// Create reader
	uri := fmt.Sprintf("redis://%s/%s", s.Addr(), testKey)
	rr, err := NewRedisReader(uri)
	if err != nil {
		t.Fatalf("NewRedisReader() error = %v", err)
	}
	defer rr.Close()

	// Test read
	ctx := context.Background()
	data, err := rr.Read(ctx)
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}

	if string(data) != testValue {
		t.Errorf("Read() got = %s, want = %s", string(data), testValue)
	}
}

func TestRedisReader_ReadNonExistentKey(t *testing.T) {
	// Setup miniredis
	s := miniredis.RunT(t)
	defer s.Close()

	// Create reader with non-existent key
	uri := fmt.Sprintf("redis://%s/non_existent_key", s.Addr())
	rr, err := NewRedisReader(uri)
	if err != nil {
		t.Fatalf("NewRedisReader() error = %v", err)
	}
	defer rr.Close()

	// Test read
	ctx := context.Background()
	_, err = rr.Read(ctx)
	if err == nil {
		t.Error("Read() expected error for non-existent key")
	}
}

func TestRedisReader_Subscribe(t *testing.T) {
	t.Skip("Skipping subscribe test as miniredis doesn't support CONFIG command for keyspace notifications")
	
	// Setup miniredis
	s := miniredis.RunT(t)
	defer s.Close()

	// Setup test data
	testKey := "config:app"
	testValue1 := `{"app": "test", "version": "1.0"}`
	testValue2 := `{"app": "test", "version": "2.0"}`
	s.Set(testKey, testValue1)

	// Create reader
	uri := fmt.Sprintf("redis://%s/%s", s.Addr(), testKey)
	rr, err := NewRedisReader(uri)
	if err != nil {
		t.Fatalf("NewRedisReader() error = %v", err)
	}
	defer rr.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Subscribe to changes
	eventChan, err := rr.Subscribe(ctx)
	if err != nil {
		t.Fatalf("Subscribe() error = %v", err)
	}

	// Wait for initial event
	select {
	case event := <-eventChan:
		if event.Error != nil {
			t.Fatalf("Initial event error = %v", event.Error)
		}
		if string(event.Data) != testValue1 {
			t.Errorf("Initial event data = %s, want = %s", string(event.Data), testValue1)
		}
	case <-ctx.Done():
		t.Fatal("Timeout waiting for initial event")
	}

	// Update the key
	go func() {
		time.Sleep(100 * time.Millisecond)
		s.Set(testKey, testValue2)
	}()

	// Wait for update event
	select {
	case event := <-eventChan:
		if event.Error != nil {
			t.Fatalf("Update event error = %v", event.Error)
		}
		if string(event.Data) != testValue2 {
			t.Errorf("Update event data = %s, want = %s", string(event.Data), testValue2)
		}
	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for update event")
	}
}

func TestRedisReader_Close(t *testing.T) {
	// Setup miniredis
	s := miniredis.RunT(t)
	defer s.Close()

	// Create reader
	uri := fmt.Sprintf("redis://%s/config:app", s.Addr())
	rr, err := NewRedisReader(uri)
	if err != nil {
		t.Fatalf("NewRedisReader() error = %v", err)
	}

	// Close reader
	err = rr.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}

	// Test that operations fail after close
	ctx := context.Background()
	_, err = rr.Read(ctx)
	if err == nil {
		t.Error("Read() should fail after Close()")
	}

	_, err = rr.Subscribe(ctx)
	if err == nil {
		t.Error("Subscribe() should fail after Close()")
	}
}

func TestParseRedisURI(t *testing.T) {
	tests := []struct {
		name     string
		uri      string
		expected *RedisConfig
		wantErr  bool
	}{
		{
			name: "basic URI",
			uri:  "redis://localhost:6379/config:app",
			expected: &RedisConfig{
				Addr: "localhost:6379",
				Key:  "config:app",
				DB:   0,
			},
			wantErr: false,
		},
		{
			name: "URI with password",
			uri:  "redis://:password@localhost:6379/config:app",
			expected: &RedisConfig{
				Addr:     "localhost:6379",
				Key:      "config:app",
				Password: "password",
				DB:       0,
			},
			wantErr: false,
		},
		{
			name: "URI with database",
			uri:  "redis://localhost:6379/config:app?db=1",
			expected: &RedisConfig{
				Addr: "localhost:6379",
				Key:  "config:app",
				DB:   1,
			},
			wantErr: false,
		},
		{
			name: "URI with timeout",
			uri:  "redis://localhost:6379/config:app?timeout=60s",
			expected: &RedisConfig{
				Addr:    "localhost:6379",
				Key:     "config:app",
				DB:      0,
				Timeout: 60 * time.Second,
			},
			wantErr: false,
		},
		{
			name:    "invalid database",
			uri:     "redis://localhost:6379/config:app?db=invalid",
			wantErr: true,
		},
		{
			name:    "negative database",
			uri:     "redis://localhost:6379/config:app?db=-1",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := reader.ParseURI(tt.uri)
			if err != nil && !tt.wantErr {
				t.Fatalf("ParseURI() error = %v", err)
			}
			if err != nil {
				return
			}

			config := &RedisConfig{
				Timeout:      DefaultTimeout,
				RetryDelay:   DefaultRetryDelay,
				MaxRetries:   DefaultRetryAttempts,
				DB:           DefaultDB,
				PoolSize:     10,
				MinIdleConns: 1,
			}

			err = parseRedisURI(u, config)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseRedisURI() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				if config.Addr != tt.expected.Addr {
					t.Errorf("Addr = %v, want %v", config.Addr, tt.expected.Addr)
				}
				if config.Key != tt.expected.Key {
					t.Errorf("Key = %v, want %v", config.Key, tt.expected.Key)
				}
				if config.DB != tt.expected.DB {
					t.Errorf("DB = %v, want %v", config.DB, tt.expected.DB)
				}
				if tt.expected.Password != "" && config.Password != tt.expected.Password {
					t.Errorf("Password = %v, want %v", config.Password, tt.expected.Password)
				}
				if tt.expected.Timeout != 0 && config.Timeout != tt.expected.Timeout {
					t.Errorf("Timeout = %v, want %v", config.Timeout, tt.expected.Timeout)
				}
			}
		})
	}
}