package json

import (
	"testing"

	"github.com/sower-proxy/conf/decoder"
)

func TestJSONDecoder_Decode(t *testing.T) {
	decoder := NewJSONDecoder()

	t.Run("valid JSON object", func(t *testing.T) {
		data := []byte(`{"name": "test", "value": 123, "enabled": true}`)
		var result map[string]any

		err := decoder.Decode(data, &result)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if result["name"] != "test" {
			t.Errorf("expected name 'test', got: %v", result["name"])
		}
		if result["value"] != float64(123) {
			t.Errorf("expected value 123, got: %v", result["value"])
		}
		if result["enabled"] != true {
			t.Errorf("expected enabled true, got: %v", result["enabled"])
		}
	})

	t.Run("valid JSON array", func(t *testing.T) {
		data := []byte(`[1, 2, 3, "test"]`)
		var result []any

		err := decoder.Decode(data, &result)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if len(result) != 4 {
			t.Errorf("expected 4 elements, got: %d", len(result))
		}
		if result[3] != "test" {
			t.Errorf("expected last element 'test', got: %v", result[3])
		}
	})

	t.Run("struct decoding", func(t *testing.T) {
		type Config struct {
			Name    string `json:"name"`
			Value   int    `json:"value"`
			Enabled bool   `json:"enabled"`
		}

		data := []byte(`{"name": "test", "value": 123, "enabled": true}`)
		var result Config

		err := decoder.Decode(data, &result)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if result.Name != "test" {
			t.Errorf("expected name 'test', got: %s", result.Name)
		}
		if result.Value != 123 {
			t.Errorf("expected value 123, got: %d", result.Value)
		}
		if result.Enabled != true {
			t.Errorf("expected enabled true, got: %v", result.Enabled)
		}
	})

	t.Run("empty data", func(t *testing.T) {
		data := []byte("")
		var result map[string]any

		err := decoder.Decode(data, &result)
		if err == nil {
			t.Fatal("expected error for empty data")
		}
	})

	t.Run("nil target", func(t *testing.T) {
		data := []byte(`{"name": "test"}`)

		err := decoder.Decode(data, nil)
		if err == nil {
			t.Fatal("expected error for nil target")
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		data := []byte(`{"name": "test", "value":}`)
		var result map[string]any

		err := decoder.Decode(data, &result)
		if err == nil {
			t.Fatal("expected error for invalid JSON")
		}
	})
}

func TestJSONDecoder_Registration(t *testing.T) {
	t.Run("format registration", func(t *testing.T) {
		dec, err := decoder.GetDecoder(FormatJSON)
		if err != nil {
			t.Fatalf("failed to get JSON decoder: %v", err)
		}
		if dec == nil {
			t.Fatal("decoder is nil")
		}
	})

	t.Run("extension mapping", func(t *testing.T) {
		format, err := decoder.FormatFromExtension(".json")
		if err != nil {
			t.Fatalf("failed to get format from extension: %v", err)
		}
		if format != FormatJSON {
			t.Errorf("expected format %s, got: %s", FormatJSON, format)
		}
	})

	t.Run("MIME type mapping", func(t *testing.T) {
		format, err := decoder.FormatFromMIME("application/json")
		if err != nil {
			t.Fatalf("failed to get format from MIME: %v", err)
		}
		if format != FormatJSON {
			t.Errorf("expected format %s, got: %s", FormatJSON, format)
		}

		format, err = decoder.FormatFromMIME("text/json")
		if err != nil {
			t.Fatalf("failed to get format from MIME: %v", err)
		}
		if format != FormatJSON {
			t.Errorf("expected format %s, got: %s", FormatJSON, format)
		}
	})
}
