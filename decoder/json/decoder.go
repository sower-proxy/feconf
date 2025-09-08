package json

import (
	"encoding/json"
	"fmt"

	"github.com/sower-proxy/conf/decoder"
)

const (
	// FormatJSON represents JSON format
	FormatJSON decoder.Format = "json"
)

// init registers JSON decoder
func init() {
	_ = decoder.RegisterDecoder(
		FormatJSON,
		NewJSONDecoder(),
		[]string{".json"},
		[]string{"application/json", "text/json"},
	)
}

// JSONDecoder implements ConfDecoder for JSON format
type JSONDecoder struct{}

// NewJSONDecoder creates a new JSON decoder
func NewJSONDecoder() *JSONDecoder {
	return &JSONDecoder{}
}

// Unmarshal decodes JSON data to target structure
func (d *JSONDecoder) Unmarshal(data []byte, v any) error {
	if len(data) == 0 {
		return fmt.Errorf("empty JSON data")
	}

	if v == nil {
		return fmt.Errorf("target cannot be nil")
	}

	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return nil
}
