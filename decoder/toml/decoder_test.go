package toml

import (
	"testing"

	"github.com/sower-proxy/conf/decoder"
)

func TestTOMLDecoder_Decode(t *testing.T) {
	decoder := NewTOMLDecoder()

	t.Run("valid TOML object", func(t *testing.T) {
		data := []byte(`name = "test"
value = 123
enabled = true`)
		var result map[string]any

		err := decoder.Unmarshal(data, &result)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if result["name"] != "test" {
			t.Errorf("expected name 'test', got: %v", result["name"])
		}
		if result["value"] != int64(123) {
			t.Errorf("expected value 123, got: %v", result["value"])
		}
		if result["enabled"] != true {
			t.Errorf("expected enabled true, got: %v", result["enabled"])
		}
	})

	t.Run("valid TOML array", func(t *testing.T) {
		data := []byte(`numbers = [1, 2, 3]
strings = ["a", "b", "c"]`)
		var result map[string]any

		err := decoder.Unmarshal(data, &result)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		numbers, ok := result["numbers"].([]any)
		if !ok {
			t.Fatalf("expected numbers to be []any, got: %T", result["numbers"])
		}
		if len(numbers) != 3 {
			t.Errorf("expected 3 numbers, got: %d", len(numbers))
		}
		if numbers[0] != int64(1) {
			t.Errorf("expected first number 1, got: %v", numbers[0])
		}

		strings, ok := result["strings"].([]any)
		if !ok {
			t.Fatalf("expected strings to be []any, got: %T", result["strings"])
		}
		if len(strings) != 3 {
			t.Errorf("expected 3 strings, got: %d", len(strings))
		}
		if strings[0] != "a" {
			t.Errorf("expected first string 'a', got: %v", strings[0])
		}
	})

	t.Run("valid TOML with sections", func(t *testing.T) {
		data := []byte(`title = "TOML Example"

[database]
host = "localhost"
port = 5432
connection_max = 5000
enabled = true

[server]
host = "127.0.0.1"
port = 8080`)
		var result map[string]any

		err := decoder.Unmarshal(data, &result)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if result["title"] != "TOML Example" {
			t.Errorf("expected title 'TOML Example', got: %v", result["title"])
		}

		dbSection, ok := result["database"].(map[string]any)
		if !ok {
			t.Fatalf("expected database section to be map[string]any, got: %T", result["database"])
		}
		if dbSection["host"] != "localhost" {
			t.Errorf("expected database host 'localhost', got: %v", dbSection["host"])
		}
		if dbSection["port"] != int64(5432) {
			t.Errorf("expected database port 5432, got: %v", dbSection["port"])
		}
		if dbSection["enabled"] != true {
			t.Errorf("expected database enabled true, got: %v", dbSection["enabled"])
		}

		serverSection, ok := result["server"].(map[string]any)
		if !ok {
			t.Fatalf("expected server section to be map[string]any, got: %T", result["server"])
		}
		if serverSection["host"] != "127.0.0.1" {
			t.Errorf("expected server host '127.0.0.1', got: %v", serverSection["host"])
		}
		if serverSection["port"] != int64(8080) {
			t.Errorf("expected server port 8080, got: %v", serverSection["port"])
		}
	})

	t.Run("struct decoding", func(t *testing.T) {
		type Config struct {
			Name    string `toml:"name"`
			Value   int    `toml:"value"`
			Enabled bool   `toml:"enabled"`
		}

		data := []byte(`name = "test"
value = 123
enabled = true`)
		var result Config

		err := decoder.Unmarshal(data, &result)
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

	t.Run("complex struct with nested data", func(t *testing.T) {
		type Database struct {
			Host string `toml:"host"`
			Port int    `toml:"port"`
		}

		type Config struct {
			Title    string   `toml:"title"`
			Database Database `toml:"database"`
		}

		data := []byte(`title = "Complex Example"

[database]
host = "localhost"
port = 5432`)
		var result Config

		err := decoder.Unmarshal(data, &result)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if result.Title != "Complex Example" {
			t.Errorf("expected title 'Complex Example', got: %s", result.Title)
		}
		if result.Database.Host != "localhost" {
			t.Errorf("expected database host 'localhost', got: %s", result.Database.Host)
		}
		if result.Database.Port != 5432 {
			t.Errorf("expected database port 5432, got: %d", result.Database.Port)
		}
	})

	t.Run("TOML with arrays of tables", func(t *testing.T) {
		data := []byte(`[[products]]
name = "Hammer"
sku = 738594937

[[products]]
name = "Nail"
sku = 284758393`)
		var result map[string]any

		err := decoder.Unmarshal(data, &result)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		products, ok := result["products"].([]map[string]any)
		if !ok {
			t.Fatalf("expected products to be []map[string]any, got: %T", result["products"])
		}
		if len(products) != 2 {
			t.Errorf("expected 2 products, got: %d", len(products))
		}

		firstProduct := products[0]
		if firstProduct["name"] != "Hammer" {
			t.Errorf("expected first product name 'Hammer', got: %v", firstProduct["name"])
		}
	})

	t.Run("empty data", func(t *testing.T) {
		data := []byte("")
		var result map[string]any

		err := decoder.Unmarshal(data, &result)
		if err == nil {
			t.Fatal("expected error for empty data")
		}
	})

	t.Run("nil target", func(t *testing.T) {
		data := []byte(`name = "test"`)

		err := decoder.Unmarshal(data, nil)
		if err == nil {
			t.Fatal("expected error for nil target")
		}
	})

	t.Run("invalid TOML", func(t *testing.T) {
		data := []byte(`[unclosed_section
name = "test"`)
		var result map[string]any

		err := decoder.Unmarshal(data, &result)
		if err == nil {
			t.Fatal("expected error for invalid TOML")
		}
	})
}

func TestTOMLDecoder_Registration(t *testing.T) {
	t.Run("format registration", func(t *testing.T) {
		dec, err := decoder.GetDecoder(FormatTOML)
		if err != nil {
			t.Fatalf("failed to get TOML decoder: %v", err)
		}
		if dec == nil {
			t.Fatal("decoder is nil")
		}
	})

	t.Run("extension mapping", func(t *testing.T) {
		format, err := decoder.FormatFromExtension(".toml")
		if err != nil {
			t.Fatalf("failed to get format from extension: %v", err)
		}
		if format != FormatTOML {
			t.Errorf("expected format %s, got: %s", FormatTOML, format)
		}
	})

	t.Run("MIME type mapping", func(t *testing.T) {
		testCases := []string{"application/toml", "text/toml"}
		for _, mime := range testCases {
			format, err := decoder.FormatFromMIME(mime)
			if err != nil {
				t.Fatalf("failed to get format from MIME %s: %v", mime, err)
			}
			if format != FormatTOML {
				t.Errorf("expected format %s for MIME %s, got: %s", FormatTOML, mime, format)
			}
		}
	})
}
