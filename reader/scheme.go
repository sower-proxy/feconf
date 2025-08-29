package reader

import (
	"fmt"
	"net/url"
	"strings"
	"sync"
)

// Scheme represents protocol type
type Scheme string

// schemeReaderMap maps scheme to reader instances using sync.Map
var schemeReaderMap sync.Map

// RegisterReader registers a reader for given scheme
func RegisterReader(scheme Scheme, reader ConfReader) error {
	if scheme == "" {
		return fmt.Errorf("empty scheme")
	}
	if reader == nil {
		return fmt.Errorf("nil reader")
	}

	schemeReaderMap.Store(scheme, reader)
	return nil
}

// GetReader returns reader for given scheme
func GetReader(scheme Scheme) (ConfReader, bool) {
	value, exists := schemeReaderMap.Load(scheme)
	if !exists {
		return nil, false
	}

	reader, ok := value.(ConfReader)
	if !ok {
		return nil, false
	}

	return reader, true
}

// ParseURI parses URI and validates configuration information
func ParseURI(uri string) (*url.URL, error) {
	if strings.TrimSpace(uri) == "" {
		return nil, fmt.Errorf("empty URI")
	}

	u, err := url.Parse(uri)
	if err != nil {
		return nil, fmt.Errorf("invalid URI format: %w", err)
	}

	if u.Scheme == "" {
		return nil, fmt.Errorf("missing scheme in URI: %s", uri)
	}

	if _, exists := schemeReaderMap.Load(Scheme(u.Scheme)); !exists {
		return nil, fmt.Errorf("unsupported URI scheme: %s", u.Scheme)
	}

	return u, nil
}
