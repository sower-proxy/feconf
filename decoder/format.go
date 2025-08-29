package decoder

import (
	"fmt"
	"strings"
	"sync"
)

// Format represents configuration file format type
type Format string

// formatDecoderMap maps format to decoder instances using sync.Map
var (
	formatDecoderMap  sync.Map
	extensionToFormat sync.Map
	mimeToFormat      sync.Map
)

// RegisterDecoder registers a decoder for given scheme
func RegisterDecoder(format Format, decoder ConfDecoder, exts []string, mimes []string) error {
	if format == "" {
		return fmt.Errorf("format cannot be empty")
	}
	if decoder == nil {
		return fmt.Errorf("decoder cannot be nil")
	}

	formatDecoderMap.Store(format, decoder)

	for _, ext := range exts {
		ext = strings.ToLower(ext)
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
		extensionToFormat.Store(ext, format)
	}

	for _, mime := range mimes {
		mime = strings.ToLower(strings.TrimSpace(mime))
		if mime != "" {
			mimeToFormat.Store(mime, format)
		}
	}

	return nil
}

// FormatFromExtension returns format from file extension with validation
func FormatFromExtension(ext string) (Format, error) {
	if strings.TrimSpace(ext) == "" {
		return "", fmt.Errorf("empty extension")
	}

	ext = strings.ToLower(ext)
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}

	value, exists := extensionToFormat.Load(ext)
	if !exists {
		return "", fmt.Errorf("unsupported file extension: %s", ext)
	}

	format, ok := value.(Format)
	if !ok {
		return "", fmt.Errorf("invalid format type for extension: %s", ext)
	}

	return format, nil
}

// FormatFromMIME returns format from MIME type with validation
func FormatFromMIME(mime string) (Format, error) {
	if strings.TrimSpace(mime) == "" {
		return "", fmt.Errorf("empty MIME type")
	}

	mime = strings.ToLower(strings.TrimSpace(mime))

	value, exists := mimeToFormat.Load(mime)
	if !exists {
		return "", fmt.Errorf("unsupported MIME type: %s", mime)
	}

	format, ok := value.(Format)
	if !ok {
		return "", fmt.Errorf("invalid format type for MIME type: %s", mime)
	}

	return format, nil
}

// GetDecoder returns decoder for given format
func GetDecoder(format Format) (ConfDecoder, error) {
	if format == "" {
		return nil, fmt.Errorf("empty format")
	}

	value, exists := formatDecoderMap.Load(format)
	if !exists {
		return nil, fmt.Errorf("no decoder registered for format: %s", format)
	}

	decoder, ok := value.(ConfDecoder)
	if !ok {
		return nil, fmt.Errorf("invalid decoder type for format: %s", format)
	}

	return decoder, nil
}
