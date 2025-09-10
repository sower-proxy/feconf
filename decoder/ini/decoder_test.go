package ini

import (
	"testing"

	"github.com/sower-proxy/feconf/decoder"
)

func TestINIDecoder_Decode(t *testing.T) {
	decoder := NewINIDecoder()

	t.Run("valid INI with default section", func(t *testing.T) {
		data := []byte(`app_name = test
debug = true
port = 8080`)
		var result map[string]any

		err := decoder.Unmarshal(data, &result)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if result["app_name"] != "test" {
			t.Errorf("expected app_name 'test', got: %v", result["app_name"])
		}
		if result["debug"] != "true" {
			t.Errorf("expected debug 'true', got: %v", result["debug"])
		}
		if result["port"] != "8080" {
			t.Errorf("expected port '8080', got: %v", result["port"])
		}
	})

	t.Run("valid INI with sections", func(t *testing.T) {
		data := []byte(`app_name = test

[database]
host = localhost
port = 5432
name = testdb

[redis]
host = redis.example.com
port = 6379`)
		var result map[string]any

		err := decoder.Unmarshal(data, &result)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if result["app_name"] != "test" {
			t.Errorf("expected app_name 'test', got: %v", result["app_name"])
		}

		dbSection, ok := result["database"].(map[string]any)
		if !ok {
			t.Fatalf("expected database section to be map[string]any, got: %T", result["database"])
		}
		if dbSection["host"] != "localhost" {
			t.Errorf("expected database host 'localhost', got: %v", dbSection["host"])
		}
		if dbSection["port"] != "5432" {
			t.Errorf("expected database port '5432', got: %v", dbSection["port"])
		}
		if dbSection["name"] != "testdb" {
			t.Errorf("expected database name 'testdb', got: %v", dbSection["name"])
		}

		redisSection, ok := result["redis"].(map[string]any)
		if !ok {
			t.Fatalf("expected redis section to be map[string]any, got: %T", result["redis"])
		}
		if redisSection["host"] != "redis.example.com" {
			t.Errorf("expected redis host 'redis.example.com', got: %v", redisSection["host"])
		}
		if redisSection["port"] != "6379" {
			t.Errorf("expected redis port '6379', got: %v", redisSection["port"])
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
		data := []byte(`app_name = test`)

		err := decoder.Unmarshal(data, nil)
		if err == nil {
			t.Fatal("expected error for nil target")
		}
	})

	t.Run("invalid INI", func(t *testing.T) {
		data := []byte(`[unclosed_section
key = value`)
		var result map[string]any

		err := decoder.Unmarshal(data, &result)
		if err == nil {
			t.Fatal("expected error for invalid INI")
		}
	})

	t.Run("unsupported target type", func(t *testing.T) {
		data := []byte(`app_name = test`)
		var result []string

		err := decoder.Unmarshal(data, &result)
		if err == nil {
			t.Fatal("expected error for unsupported target type")
		}
	})

	t.Run("map target without pointer", func(t *testing.T) {
		data := []byte(`app_name = test
port = 8080`)
		result := make(map[string]any)

		err := decoder.Unmarshal(data, result)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if result["app_name"] != "test" {
			t.Errorf("expected app_name 'test', got: %v", result["app_name"])
		}
		if result["port"] != "8080" {
			t.Errorf("expected port '8080', got: %v", result["port"])
		}
	})
}

func TestINIDecoder_Registration(t *testing.T) {
	t.Run("format registration", func(t *testing.T) {
		dec, err := decoder.GetDecoder(FormatINI)
		if err != nil {
			t.Fatalf("failed to get INI decoder: %v", err)
		}
		if dec == nil {
			t.Fatal("decoder is nil")
		}
	})

	t.Run("extension mapping", func(t *testing.T) {
		testCases := []string{".ini", ".cfg", ".conf"}
		for _, ext := range testCases {
			format, err := decoder.FormatFromExtension(ext)
			if err != nil {
				t.Fatalf("failed to get format from extension %s: %v", ext, err)
			}
			if format != FormatINI {
				t.Errorf("expected format %s for extension %s, got: %s", FormatINI, ext, format)
			}
		}
	})

	t.Run("MIME type mapping", func(t *testing.T) {
		testCases := []string{"text/ini", "application/ini"}
		for _, mime := range testCases {
			format, err := decoder.FormatFromMIME(mime)
			if err != nil {
				t.Fatalf("failed to get format from MIME %s: %v", mime, err)
			}
			if format != FormatINI {
				t.Errorf("expected format %s for MIME %s, got: %s", FormatINI, mime, format)
			}
		}
	})
}
