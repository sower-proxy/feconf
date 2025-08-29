package decoder

// ConfDecoder defines configuration decoder interface
type ConfDecoder interface {
	// Decode decodes raw configuration data to target structure
	Decode(data []byte, v any) error
}
