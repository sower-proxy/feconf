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
func RegisterReader(scheme Scheme, newReader func(uri string) (ConfReader, error)) error {
	if scheme == "" {
		return fmt.Errorf("empty scheme")
	}
	if newReader == nil {
		return fmt.Errorf("nil reader")
	}

	schemeReaderMap.Store(scheme, newReader)
	return nil
}

// GetReader creates and returns reader for given scheme and URI
func GetReader(scheme Scheme, uri string) (ConfReader, error) {
	value, exists := schemeReaderMap.Load(scheme)
	if !exists {
		return nil, fmt.Errorf("unsupported scheme: %s", scheme)
	}

	newReaderFunc, ok := value.(func(uri string) (ConfReader, error))
	if !ok {
		return nil, fmt.Errorf("invalid reader constructor for scheme: %s", scheme)
	}

	return newReaderFunc(uri)
}

// NewReader creates reader from URI by detecting scheme automatically
func NewReader(uri string) (ConfReader, error) {
	u, err := ParseURI(uri)
	if err != nil {
		return nil, err
	}

	return GetReader(Scheme(u.Scheme), uri)
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
