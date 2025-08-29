package hcl

import (
	"fmt"

	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/sower-proxy/conf/decoder"
)

const (
	// FormatHCL represents HCL format
	FormatHCL decoder.Format = "hcl"
)

// init registers HCL decoder
func init() {
	decoder.RegisterDecoder(
		FormatHCL,
		NewHCLDecoder(),
		[]string{".hcl", ".tf"},
		[]string{"application/hcl", "text/hcl"},
	)
}

// HCLDecoder implements ConfDecoder for HCL format
type HCLDecoder struct{}

// NewHCLDecoder creates a new HCL decoder
func NewHCLDecoder() *HCLDecoder {
	return &HCLDecoder{}
}

// Unmarshal decodes HCL data to target structure
func (d *HCLDecoder) Unmarshal(data []byte, v any) error {
	if len(data) == 0 {
		return fmt.Errorf("empty HCL data")
	}

	if v == nil {
		return fmt.Errorf("target cannot be nil")
	}

	// HCL v2 requires a filename for parsing context
	filename := "config.hcl"

	// Use hclsimple to decode HCL data directly to the target
	err := hclsimple.Decode(filename, data, nil, v)
	if err != nil {
		return fmt.Errorf("failed to unmarshal HCL: %w", err)
	}

	return nil
}
