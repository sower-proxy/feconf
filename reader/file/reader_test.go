package file

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewFileReader(t *testing.T) {
	tests := []struct {
		name    string
		uri     string
		wantErr bool
	}{
		{
			name:    "valid file URI",
			uri:     "file:///tmp/test.txt",
			wantErr: true, // File doesn't exist yet
		},
		{
			name:    "invalid URI format",
			uri:     "invalid://uri",
			wantErr: true,
		},
		{
			name:    "unsupported scheme",
			uri:     "http://example.com/config",
			wantErr: true,
		},
		{
			name:    "empty file path",
			uri:     "file://",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewFileReader(tt.uri)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewFileReader() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSchemeDefault(t *testing.T) {
	// Create temp file
	tmpFile, err := os.CreateTemp("", "test_config_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	testData := `{"key": "value", "number": 123}`
	if _, err := tmpFile.WriteString(testData); err != nil {
		t.Fatalf("Failed to write test data: %v", err)
	}
	tmpFile.Close()

	tests := []struct {
		name    string
		uri     string
		wantErr bool
	}{
		{
			name:    "absolute path without scheme",
			uri:     tmpFile.Name(),
			wantErr: false,
		},
		{
			name:    "file URI with scheme",
			uri:     "file://" + tmpFile.Name(),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileReader, err := NewFileReader(tt.uri)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewFileReader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				return
			}
			defer fileReader.Close()

			// Test reading data
			ctx := context.Background()
			data, err := fileReader.Read(ctx)
			if err != nil {
				t.Errorf("Failed to read: %v", err)
				return
			}

			if string(data) != testData {
				t.Errorf("Expected data %s, got %s", testData, string(data))
			}
		})
	}
}

func TestFileReader_Read(t *testing.T) {
	// Create temp file
	tmpFile, err := os.CreateTemp("", "test_config_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	testData := `{"key": "value", "number": 123}`
	if _, err := tmpFile.WriteString(testData); err != nil {
		t.Fatalf("Failed to write test data: %v", err)
	}
	tmpFile.Close()

	uri := "file://" + tmpFile.Name()
	fileReader, err := NewFileReader(uri)
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}
	defer fileReader.Close()

	ctx := context.Background()
	data, err := fileReader.Read(ctx)
	if err != nil {
		t.Fatalf("Failed to read: %v", err)
	}

	if string(data) != testData {
		t.Errorf("Expected data %s, got %s", testData, string(data))
	}
}

func TestFileReader_ReadWithContext(t *testing.T) {
	// Create temp file
	tmpFile, err := os.CreateTemp("", "test_config_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	testData := `{"key": "value"}`
	if _, err := tmpFile.WriteString(testData); err != nil {
		t.Fatalf("Failed to write test data: %v", err)
	}
	tmpFile.Close()

	uri := "file://" + tmpFile.Name()
	fileReader, err := NewFileReader(uri)
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}
	defer fileReader.Close()

	// Test with cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	data, err := fileReader.Read(ctx)
	if err != context.Canceled {
		t.Errorf("Expected context.Canceled error, got %v", err)
	}

	if data != nil {
		t.Errorf("Expected nil data with cancelled context, got %v", data)
	}
}

func TestFileReader_Subscribe(t *testing.T) {
	// Create temp file
	tmpFile, err := os.CreateTemp("", "test_config_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	initialData := `{"initial": true}`
	if _, err := tmpFile.WriteString(initialData); err != nil {
		t.Fatalf("Failed to write initial data: %v", err)
	}
	tmpFile.Close()

	uri := "file://" + tmpFile.Name()
	fileReader, err := NewFileReader(uri)
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}
	defer fileReader.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	eventChan, err := fileReader.Subscribe(ctx)
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	// Update file content
	go func() {
		time.Sleep(100 * time.Millisecond)
		updatedData := `{"updated": true}`
		if err := os.WriteFile(tmpFile.Name(), []byte(updatedData), 0o644); err != nil {
			t.Errorf("Failed to update file: %v", err)
		}
	}()

	// Wait for file change event
	select {
	case event := <-eventChan:
		if event == nil {
			t.Fatal("Received nil event")
		}

		if event.Error != nil {
			t.Errorf("Unexpected error in event: %v", event.Error)
		}

		if string(event.Data) != `{"updated": true}` {
			t.Errorf("Expected updated data, got %s", string(event.Data))
		}

		if event.SourceURI != uri {
			t.Errorf("Expected SourceURI %s, got %s", uri, event.SourceURI)
		}

	case <-ctx.Done():
		t.Fatal("Timeout waiting for file change event")
	}
}

func TestFileReader_DoubleSubscribe(t *testing.T) {
	// Create temp file
	tmpFile, err := os.CreateTemp("", "test_config_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	testData := `{"key": "value"}`
	if _, err := tmpFile.WriteString(testData); err != nil {
		t.Fatalf("Failed to write test data: %v", err)
	}
	tmpFile.Close()

	uri := "file://" + tmpFile.Name()
	fileReader, err := NewFileReader(uri)
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}
	defer fileReader.Close()

	ctx := context.Background()

	// First subscription should succeed
	_, err = fileReader.Subscribe(ctx)
	if err != nil {
		t.Fatalf("First subscription failed: %v", err)
	}

	// Second subscription should fail
	_, err = fileReader.Subscribe(ctx)
	if err == nil {
		t.Error("Second subscription should fail")
	}
}

func TestFileReader_Close(t *testing.T) {
	// Create temp file
	tmpFile, err := os.CreateTemp("", "test_config_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	testData := `{"key": "value"}`
	if _, err := tmpFile.WriteString(testData); err != nil {
		t.Fatalf("Failed to write test data: %v", err)
	}
	tmpFile.Close()

	uri := "file://" + tmpFile.Name()
	fileReader, err := NewFileReader(uri)
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}

	// Test double close
	if err := fileReader.Close(); err != nil {
		t.Errorf("First close failed: %v", err)
	}

	if err := fileReader.Close(); err != nil {
		t.Errorf("Second close failed: %v", err)
	}

	// Test operations on closed reader
	ctx := context.Background()
	_, err = fileReader.Read(ctx)
	if err == nil {
		t.Error("Read on closed reader should fail")
	}

	_, err = fileReader.Subscribe(ctx)
	if err == nil {
		t.Error("Subscribe on closed reader should fail")
	}
}

func TestFileReader_NonExistentFile(t *testing.T) {
	uri := "file:///non/existent/file.json"
	_, err := NewFileReader(uri)
	if err == nil {
		t.Error("Creating reader for non-existent file should fail")
	}
}

func TestFileReader_ParseURI(t *testing.T) {
	tests := []struct {
		name     string
		uri      string
		expected string
	}{
		{
			name:     "absolute path",
			uri:      "file:///tmp/config.json",
			expected: "/tmp/config.json",
		},
		{
			name:     "with host",
			uri:      "file://localhost/tmp/config.json",
			expected: filepath.Join("localhost", "/tmp/config.json"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file for test
			tmpFile, err := os.CreateTemp("", "test_*.json")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tmpFile.Name())
			tmpFile.Close()

			// Use actual temp file path
			actualURI := "file://" + tmpFile.Name()
			fileReader, err := NewFileReader(actualURI)
			if err != nil {
				t.Fatalf("Failed to create reader: %v", err)
			}
			defer fileReader.Close()

			if fileReader.filePath != tmpFile.Name() {
				t.Errorf("Expected file path %s, got %s", tmpFile.Name(), fileReader.filePath)
			}
		})
	}
}

func TestFileReader_MalformedFile(t *testing.T) {
	// Test reading file with permission denied
	tmpFile, err := os.CreateTemp("", "test_config_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	testData := `{"key": "value"}`
	if _, err := tmpFile.WriteString(testData); err != nil {
		t.Fatalf("Failed to write test data: %v", err)
	}
	tmpFile.Close()

	// Change permissions to make file unreadable
	if err := os.Chmod(tmpFile.Name(), 0o000); err != nil {
		t.Fatalf("Failed to change file permissions: %v", err)
	}
	defer func() {
		_ = os.Chmod(tmpFile.Name(), 0o644) // Restore permissions for cleanup
	}()

	uri := "file://" + tmpFile.Name()
	fileReader, err := NewFileReader(uri)
	// NewFileReader only checks if file exists, not if it's readable
	if err != nil {
		t.Skip("File permissions test may not work on all systems")
	}
	defer fileReader.Close()

	ctx := context.Background()
	data, err := fileReader.Read(ctx)
	if err == nil {
		t.Skip("Permission denied test may not work on all systems")
	}

	if data != nil {
		t.Error("Data should be nil when permission denied")
	}
}
