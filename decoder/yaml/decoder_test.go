package yaml

import (
	"testing"

	"github.com/sower-proxy/feconf/decoder"
)

func TestYAMLDecoder_Decode(t *testing.T) {
	decoder := NewYAMLDecoder()

	t.Run("valid YAML object", func(t *testing.T) {
		data := []byte(`name: test
value: 123
enabled: true`)
		var result map[string]any

		err := decoder.Unmarshal(data, &result)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if result["name"] != "test" {
			t.Errorf("expected name 'test', got: %v", result["name"])
		}
		if result["value"] != 123 {
			t.Errorf("expected value 123, got: %v", result["value"])
		}
		if result["enabled"] != true {
			t.Errorf("expected enabled true, got: %v", result["enabled"])
		}
	})

	t.Run("valid YAML array", func(t *testing.T) {
		data := []byte(`- 1
- 2
- 3
- test`)
		var result []any

		err := decoder.Unmarshal(data, &result)
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

	t.Run("nested YAML structure", func(t *testing.T) {
		data := []byte(`database:
  host: localhost
  port: 5432
  credentials:
    username: user
    password: pass
features:
  - auth
  - logging
  - metrics`)
		var result map[string]any

		err := decoder.Unmarshal(data, &result)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		db, ok := result["database"].(map[string]any)
		if !ok {
			t.Fatal("expected database to be a map")
		}
		if db["host"] != "localhost" {
			t.Errorf("expected host 'localhost', got: %v", db["host"])
		}
		if db["port"] != 5432 {
			t.Errorf("expected port 5432, got: %v", db["port"])
		}

		creds, ok := db["credentials"].(map[string]any)
		if !ok {
			t.Fatal("expected credentials to be a map")
		}
		if creds["username"] != "user" {
			t.Errorf("expected username 'user', got: %v", creds["username"])
		}

		features, ok := result["features"].([]any)
		if !ok {
			t.Fatal("expected features to be an array")
		}
		if len(features) != 3 {
			t.Errorf("expected 3 features, got: %d", len(features))
		}
		if features[0] != "auth" {
			t.Errorf("expected first feature 'auth', got: %v", features[0])
		}
	})

	t.Run("struct decoding", func(t *testing.T) {
		type Config struct {
			Name    string `yaml:"name"`
			Value   int    `yaml:"value"`
			Enabled bool   `yaml:"enabled"`
		}

		data := []byte(`name: test
value: 123
enabled: true`)
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
			Host string `yaml:"host"`
			Port int    `yaml:"port"`
		}

		type Config struct {
			Database Database `yaml:"database"`
			Features []string `yaml:"features"`
			Debug    bool     `yaml:"debug"`
		}

		data := []byte(`database:
  host: localhost
  port: 5432
features:
  - auth
  - logging
debug: true`)
		var result Config

		err := decoder.Unmarshal(data, &result)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if result.Database.Host != "localhost" {
			t.Errorf("expected host 'localhost', got: %s", result.Database.Host)
		}
		if result.Database.Port != 5432 {
			t.Errorf("expected port 5432, got: %d", result.Database.Port)
		}
		if len(result.Features) != 2 {
			t.Errorf("expected 2 features, got: %d", len(result.Features))
		}
		if result.Features[0] != "auth" {
			t.Errorf("expected first feature 'auth', got: %s", result.Features[0])
		}
		if result.Debug != true {
			t.Errorf("expected debug true, got: %v", result.Debug)
		}
	})

	t.Run("YAML with comments", func(t *testing.T) {
		data := []byte(`# Configuration file
name: test  # application name
value: 123  # numeric value
# Enable feature flag
enabled: true`)
		var result map[string]any

		err := decoder.Unmarshal(data, &result)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if result["name"] != "test" {
			t.Errorf("expected name 'test', got: %v", result["name"])
		}
		if result["value"] != 123 {
			t.Errorf("expected value 123, got: %v", result["value"])
		}
	})

	t.Run("YAML multiline strings", func(t *testing.T) {
		data := []byte(`description: |
  This is a multiline
  description that spans
  multiple lines.
compact: >
  This is a folded
  string that will be
  on one line.`)
		var result map[string]any

		err := decoder.Unmarshal(data, &result)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		desc, ok := result["description"].(string)
		if !ok {
			t.Fatal("expected description to be a string")
		}
		if desc != "This is a multiline\ndescription that spans\nmultiple lines.\n" {
			t.Errorf("unexpected multiline string: %q", desc)
		}

		compact, ok := result["compact"].(string)
		if !ok {
			t.Fatal("expected compact to be a string")
		}
		if compact != "This is a folded string that will be on one line." {
			t.Errorf("unexpected folded string: %q", compact)
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
		data := []byte(`name: test`)

		err := decoder.Unmarshal(data, nil)
		if err == nil {
			t.Fatal("expected error for nil target")
		}
	})

	t.Run("invalid YAML", func(t *testing.T) {
		data := []byte(`name: test
  value: 123
invalid: [unclosed`)
		var result map[string]any

		err := decoder.Unmarshal(data, &result)
		if err == nil {
			t.Fatal("expected error for invalid YAML")
		}
	})

	t.Run("YAML with anchors and aliases", func(t *testing.T) {
		data := []byte(`default: &default
  host: localhost
  port: 5432

development:
  <<: *default
  database: dev_db

production:
  <<: *default
  database: prod_db
  port: 5433`)
		var result map[string]any

		err := decoder.Unmarshal(data, &result)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		dev, ok := result["development"].(map[string]any)
		if !ok {
			t.Fatal("expected development to be a map")
		}
		if dev["host"] != "localhost" {
			t.Errorf("expected dev host 'localhost', got: %v", dev["host"])
		}
		if dev["database"] != "dev_db" {
			t.Errorf("expected dev database 'dev_db', got: %v", dev["database"])
		}

		prod, ok := result["production"].(map[string]any)
		if !ok {
			t.Fatal("expected production to be a map")
		}
		if prod["host"] != "localhost" {
			t.Errorf("expected prod host 'localhost', got: %v", prod["host"])
		}
		if prod["port"] != 5433 {
			t.Errorf("expected prod port 5433, got: %v", prod["port"])
		}
	})
}

func TestYAMLDecoder_Registration(t *testing.T) {
	t.Run("format registration", func(t *testing.T) {
		dec, err := decoder.GetDecoder(FormatYAML)
		if err != nil {
			t.Fatalf("failed to get YAML decoder: %v", err)
		}
		if dec == nil {
			t.Fatal("decoder is nil")
		}
	})

	t.Run("extension mapping", func(t *testing.T) {
		// Test .yaml extension
		format, err := decoder.FormatFromExtension(".yaml")
		if err != nil {
			t.Fatalf("failed to get format from .yaml extension: %v", err)
		}
		if format != FormatYAML {
			t.Errorf("expected format %s, got: %s", FormatYAML, format)
		}

		// Test .yml extension
		format, err = decoder.FormatFromExtension(".yml")
		if err != nil {
			t.Fatalf("failed to get format from .yml extension: %v", err)
		}
		if format != FormatYAML {
			t.Errorf("expected format %s, got: %s", FormatYAML, format)
		}

		// Test case insensitive
		format, err = decoder.FormatFromExtension(".YML")
		if err != nil {
			t.Fatalf("failed to get format from .YML extension: %v", err)
		}
		if format != FormatYAML {
			t.Errorf("expected format %s, got: %s", FormatYAML, format)
		}
	})

	t.Run("MIME type mapping", func(t *testing.T) {
		mimeTypes := []string{
			"application/yaml",
			"application/x-yaml",
			"text/yaml",
			"text/x-yaml",
		}

		for _, mime := range mimeTypes {
			format, err := decoder.FormatFromMIME(mime)
			if err != nil {
				t.Fatalf("failed to get format from MIME %s: %v", mime, err)
			}
			if format != FormatYAML {
				t.Errorf("expected format %s for MIME %s, got: %s", FormatYAML, mime, format)
			}
		}
	})
}

func TestYAMLDecoder_EdgeCases(t *testing.T) {
	decoder := NewYAMLDecoder()

	t.Run("YAML with different data types", func(t *testing.T) {
		data := []byte(`string: "hello"
integer: 42
float: 3.14
boolean: true
null_value: null
date: 2023-01-01
list:
  - item1
  - item2
nested:
  key: value`)
		var result map[string]any

		err := decoder.Unmarshal(data, &result)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if result["string"] != "hello" {
			t.Errorf("expected string 'hello', got: %v", result["string"])
		}
		if result["integer"] != 42 {
			t.Errorf("expected integer 42, got: %v", result["integer"])
		}
		if result["float"] != 3.14 {
			t.Errorf("expected float 3.14, got: %v", result["float"])
		}
		if result["boolean"] != true {
			t.Errorf("expected boolean true, got: %v", result["boolean"])
		}
		if result["null_value"] != nil {
			t.Errorf("expected null value nil, got: %v", result["null_value"])
		}
	})

	t.Run("empty YAML document", func(t *testing.T) {
		data := []byte(`---`)
		var result map[string]any

		err := decoder.Unmarshal(data, &result)
		if err != nil {
			t.Fatalf("expected no error for empty YAML document, got: %v", err)
		}

		if result != nil {
			t.Errorf("expected nil result for empty YAML document, got: %v", result)
		}
	})

	t.Run("multiple YAML documents", func(t *testing.T) {
		data := []byte(`name: first
---
name: second`)
		var result map[string]any

		err := decoder.Unmarshal(data, &result)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		// yaml.Unmarshal only processes the first document
		if result["name"] != "first" {
			t.Errorf("expected name 'first', got: %v", result["name"])
		}
	})
}
