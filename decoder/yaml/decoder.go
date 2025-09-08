package yaml

import (
	"fmt"

	"github.com/sower-proxy/conf/decoder"
	"gopkg.in/yaml.v3"
)

const (
	// FormatYAML represents YAML format
	FormatYAML decoder.Format = "yaml"
)

// init registers YAML decoder
func init() {
	_ = decoder.RegisterDecoder(
		FormatYAML,
		NewYAMLDecoder(),
		[]string{".yaml", ".yml"},
		[]string{"application/yaml", "application/x-yaml", "text/yaml", "text/x-yaml"},
	)
}

// YAMLDecoder implements ConfDecoder for YAML format
type YAMLDecoder struct{}

// NewYAMLDecoder creates a new YAML decoder
func NewYAMLDecoder() *YAMLDecoder {
	return &YAMLDecoder{}
}

// Unmarshal decodes YAML data to target structure
func (d *YAMLDecoder) Unmarshal(data []byte, v any) error {
	if len(data) == 0 {
		return fmt.Errorf("empty YAML data")
	}

	if v == nil {
		return fmt.Errorf("target cannot be nil")
	}

	if err := yaml.Unmarshal(data, v); err != nil {
		return fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	return nil
}
