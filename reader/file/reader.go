package file

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/sower-proxy/feconf/reader"
)

const (
	// SchemeFile represents file URI scheme
	SchemeFile    reader.Scheme = "file"
	SchemeDefault reader.Scheme = ""
)

// init registers file reader
func init() {
	_ = reader.RegisterReader(SchemeFile, func(uri string) (reader.ConfReader, error) {
		return NewFileReader(uri)
	})
	_ = reader.RegisterReader(SchemeDefault, func(uri string) (reader.ConfReader, error) {
		return NewFileReader(uri)
	})
}

// FileReader implements ConfReader for file-based configuration
type FileReader struct {
	uri      string
	filePath string
	watcher  *fsnotify.Watcher
	mu       sync.RWMutex
	closed   bool
}

// NewFileReader creates a new file reader
func NewFileReader(uri string) (*FileReader, error) {
	u, err := reader.ParseURI(uri)
	if err != nil {
		return nil, err
	}

	if u.Scheme != string(SchemeFile) && u.Scheme != string(SchemeDefault) {
		return nil, fmt.Errorf("unsupported scheme: %s, expected: %s or empty", u.Scheme, SchemeFile)
	}

	filePath := u.Path
	if u.Host != "" {
		filePath = filepath.Join(u.Host, u.Path)
	}

	// Validate file path
	if filePath == "" {
		return nil, fmt.Errorf("empty file path")
	}

	// Check if file exists and is readable
	if _, err := os.Stat(filePath); err != nil {
		return nil, fmt.Errorf("file access error: %w", err)
	}

	return &FileReader{
		uri:      uri,
		filePath: filePath,
	}, nil
}

// Read reads configuration data from file
func (f *FileReader) Read(ctx context.Context) ([]byte, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if f.closed {
		return nil, fmt.Errorf("reader is closed")
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	return f.readFile()
}

// Subscribe subscribes to file changes and returns update channel
func (f *FileReader) Subscribe(ctx context.Context) (<-chan *reader.ReadEvent, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.closed {
		return nil, fmt.Errorf("reader is closed")
	}

	if f.watcher != nil {
		return nil, fmt.Errorf("already subscribed")
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %w", err)
	}

	// Watch the file
	if err := watcher.Add(f.filePath); err != nil {
		_ = watcher.Close()
		return nil, fmt.Errorf("failed to watch file %s: %w", f.filePath, err)
	}

	f.watcher = watcher
	eventChan := make(chan *reader.ReadEvent, 1)

	go f.watchFile(ctx, eventChan)

	return eventChan, nil
}

// Close closes the reader and cleans up resources
func (f *FileReader) Close() error {
	if f == nil {
		return nil
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	if f.closed {
		return nil
	}

	f.closed = true

	if f.watcher != nil {
		if err := f.watcher.Close(); err != nil {
			return fmt.Errorf("failed to close file watcher: %w", err)
		}
		f.watcher = nil
	}

	return nil
}

// readFile reads file content with proper error handling
func (f *FileReader) readFile() ([]byte, error) {
	file, err := os.Open(f.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", f.filePath, err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", f.filePath, err)
	}

	return data, nil
}

// watchFile watches for file changes and sends events
func (f *FileReader) watchFile(ctx context.Context, eventChan chan<- *reader.ReadEvent) {
	defer close(eventChan)

	for {
		f.mu.RLock()
		if f.closed || f.watcher == nil {
			f.mu.RUnlock()
			return
		}

		watcher := f.watcher
		f.mu.RUnlock()

		select {
		case <-ctx.Done():
			return
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			// Handle file write/modify events
			if event.Op&fsnotify.Write == fsnotify.Write ||
				event.Op&fsnotify.Create == fsnotify.Create {

				// Add small delay to ensure file write is complete
				time.Sleep(10 * time.Millisecond)

				data, err := f.readFile()
				confEvent := reader.NewReadEvent(f.uri, data, err)

				select {
				case eventChan <- confEvent:
				case <-ctx.Done():
					return
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}

			confEvent := reader.NewReadEvent(f.uri, nil, fmt.Errorf("file watcher error: %w", err))
			select {
			case eventChan <- confEvent:
			case <-ctx.Done():
				return
			}
		}
	}
}
