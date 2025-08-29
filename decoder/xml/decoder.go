package xml

import (
	"encoding/xml"
	"fmt"

	"github.com/sower-proxy/conf/decoder"
)

const (
	// FormatXML represents XML format
	FormatXML decoder.Format = "xml"
)

// init registers XML decoder
func init() {
	decoder.RegisterDecoder(
		FormatXML,
		NewXMLDecoder(),
		[]string{".xml"},
		[]string{"application/xml", "text/xml"},
	)
}

// XMLDecoder implements ConfDecoder for XML format
type XMLDecoder struct{}

// NewXMLDecoder creates a new XML decoder
func NewXMLDecoder() *XMLDecoder {
	return &XMLDecoder{}
}

// Unmarshal decodes XML data to target structure
func (d *XMLDecoder) Unmarshal(data []byte, v any) error {
	if len(data) == 0 {
		return fmt.Errorf("empty XML data")
	}

	if v == nil {
		return fmt.Errorf("target cannot be nil")
	}

	if err := xml.Unmarshal(data, v); err != nil {
		return fmt.Errorf("failed to unmarshal XML: %w", err)
	}

	return nil
}
