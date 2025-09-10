package toml

import (
	"fmt"

	"github.com/BurntSushi/toml"
	"github.com/sower-proxy/feconf/decoder"
)

const (
	// FormatTOML represents TOML format
	FormatTOML decoder.Format = "toml"
)

// init registers TOML decoder
func init() {
	_ = decoder.RegisterDecoder(
		FormatTOML,
		NewTOMLDecoder(),
		[]string{".toml"},
		[]string{"application/toml", "text/toml"},
	)
}

// TOMLDecoder implements ConfDecoder for TOML format
type TOMLDecoder struct{}

// NewTOMLDecoder creates a new TOML decoder
func NewTOMLDecoder() *TOMLDecoder {
	return &TOMLDecoder{}
}

// Unmarshal decodes TOML data to target structure
func (d *TOMLDecoder) Unmarshal(data []byte, v any) error {
	if len(data) == 0 {
		return fmt.Errorf("empty TOML data")
	}

	if v == nil {
		return fmt.Errorf("target cannot be nil")
	}

	if err := toml.Unmarshal(data, v); err != nil {
		return fmt.Errorf("failed to unmarshal TOML: %w", err)
	}

	return nil
}
