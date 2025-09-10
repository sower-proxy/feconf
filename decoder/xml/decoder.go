package xml

import (
	"encoding/xml"
	"fmt"

	"github.com/sower-proxy/feconf/decoder"
)

const (
	// FormatXML represents XML format
	FormatXML decoder.Format = "xml"
)

// init registers XML decoder
func init() {
	_ = decoder.RegisterDecoder(
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

	// Handle map[string]any for configuration library compatibility
	if mapTarget, ok := v.(*map[string]any); ok {
		// For XML, we need to parse into a generic structure first
		var doc xmlDoc
		if err := xml.Unmarshal(data, &doc); err != nil {
			return fmt.Errorf("failed to unmarshal XML to generic structure: %w", err)
		}
		*mapTarget = doc.toMap()
		return nil
	}

	if err := xml.Unmarshal(data, v); err != nil {
		return fmt.Errorf("failed to unmarshal XML: %w", err)
	}

	return nil
}

// xmlDoc represents a generic XML document structure
type xmlDoc struct {
	XMLName xml.Name
	Attrs   []xml.Attr `xml:",attr"`
	Content []byte     `xml:",innerxml"`
	Nodes   []xmlNode  `xml:",any"`
}

// xmlNode represents a generic XML node
type xmlNode struct {
	XMLName xml.Name
	Attrs   []xml.Attr `xml:",attr"`
	Content string     `xml:",chardata"`
	Nodes   []xmlNode  `xml:",any"`
}

// toMap converts xmlDoc to map[string]any
func (doc *xmlDoc) toMap() map[string]any {
	result := make(map[string]any)

	// Add attributes if any
	for _, attr := range doc.Attrs {
		result["@"+attr.Name.Local] = attr.Value
	}

	// Process nodes
	nodeMap := make(map[string][]xmlNode)
	for _, node := range doc.Nodes {
		key := node.XMLName.Local
		nodeMap[key] = append(nodeMap[key], node)
	}

	for key, nodes := range nodeMap {
		if len(nodes) == 1 {
			result[key] = nodes[0].toValue()
		} else {
			values := make([]any, len(nodes))
			for i, node := range nodes {
				values[i] = node.toValue()
			}
			result[key] = values
		}
	}

	return result
}

// toValue converts xmlNode to appropriate value
func (node *xmlNode) toValue() any {
	// If has child nodes, convert to map
	if len(node.Nodes) > 0 {
		result := make(map[string]any)

		// Add attributes if any
		for _, attr := range node.Attrs {
			result["@"+attr.Name.Local] = attr.Value
		}

		// Process child nodes
		nodeMap := make(map[string][]xmlNode)
		for _, child := range node.Nodes {
			key := child.XMLName.Local
			nodeMap[key] = append(nodeMap[key], child)
		}

		for key, children := range nodeMap {
			if len(children) == 1 {
				result[key] = children[0].toValue()
			} else {
				values := make([]any, len(children))
				for i, child := range children {
					values[i] = child.toValue()
				}
				result[key] = values
			}
		}

		return result
	}

	// Otherwise return content as string
	return node.Content
}
