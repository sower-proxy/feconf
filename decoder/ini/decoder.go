package ini

import (
	"fmt"

	"github.com/sower-proxy/conf/decoder"
	"gopkg.in/ini.v1"
)

const (
	// FormatINI represents INI format
	FormatINI decoder.Format = "ini"
)

// init registers INI decoder
func init() {
	_ = decoder.RegisterDecoder(
		FormatINI,
		NewINIDecoder(),
		[]string{".ini", ".cfg", ".conf"},
		[]string{"text/ini", "application/ini"},
	)
}

// INIDecoder implements ConfDecoder for INI format
type INIDecoder struct{}

// NewINIDecoder creates a new INI decoder
func NewINIDecoder() *INIDecoder {
	return &INIDecoder{}
}

// Unmarshal decodes INI data to target structure
func (d *INIDecoder) Unmarshal(data []byte, v any) error {
	if len(data) == 0 {
		return fmt.Errorf("empty INI data")
	}

	if v == nil {
		return fmt.Errorf("target cannot be nil")
	}

	// Load INI data
	cfg, err := ini.Load(data)
	if err != nil {
		return fmt.Errorf("failed to parse INI: %w", err)
	}

	// Convert to map[string]any
	result := make(map[string]any)

	// Process each section
	for _, section := range cfg.Sections() {
		sectionName := section.Name()

		// Handle default section
		if sectionName == ini.DefaultSection {
			// Add keys from default section to root
			for _, key := range section.Keys() {
				result[key.Name()] = key.String()
			}
		} else {
			// Create section map
			sectionMap := make(map[string]any)
			for _, key := range section.Keys() {
				sectionMap[key.Name()] = key.String()
			}
			result[sectionName] = sectionMap
		}
	}

	// Convert result to target type
	switch target := v.(type) {
	case *map[string]any:
		*target = result
	case map[string]any:
		for k, val := range result {
			target[k] = val
		}
	default:
		return fmt.Errorf("unsupported target type: %T", v)
	}

	return nil
}
