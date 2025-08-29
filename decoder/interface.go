package decoder

// ConfDecoder defines configuration decoder interface
type ConfDecoder interface {
	Unmarshal(data []byte, v any) error
}
