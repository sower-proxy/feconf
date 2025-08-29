package hcl

import (
	"testing"

	"github.com/sower-proxy/conf/decoder"
)

func TestHCLDecoder_Decode(t *testing.T) {
	decoder := NewHCLDecoder()

	t.Run("valid HCL with simple values", func(t *testing.T) {
		data := []byte(`name = "test"
value = 123
enabled = true`)

		type Config struct {
			Name    string `hcl:"name"`
			Value   int    `hcl:"value"`
			Enabled bool   `hcl:"enabled"`
		}
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

	t.Run("valid HCL with blocks", func(t *testing.T) {
		data := []byte(`title = "HCL Example"

database {
  host = "localhost"
  port = 5432
  enabled = true
}

server {
  host = "127.0.0.1"
  port = 8080
}`)

		type Database struct {
			Host    string `hcl:"host"`
			Port    int    `hcl:"port"`
			Enabled bool   `hcl:"enabled"`
		}

		type Server struct {
			Host string `hcl:"host"`
			Port int    `hcl:"port"`
		}

		type Config struct {
			Title    string   `hcl:"title"`
			Database Database `hcl:"database,block"`
			Server   Server   `hcl:"server,block"`
		}
		var result Config

		err := decoder.Unmarshal(data, &result)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if result.Title != "HCL Example" {
			t.Errorf("expected title 'HCL Example', got: %s", result.Title)
		}
		if result.Database.Host != "localhost" {
			t.Errorf("expected database host 'localhost', got: %s", result.Database.Host)
		}
		if result.Database.Port != 5432 {
			t.Errorf("expected database port 5432, got: %d", result.Database.Port)
		}
		if result.Database.Enabled != true {
			t.Errorf("expected database enabled true, got: %v", result.Database.Enabled)
		}
		if result.Server.Host != "127.0.0.1" {
			t.Errorf("expected server host '127.0.0.1', got: %s", result.Server.Host)
		}
		if result.Server.Port != 8080 {
			t.Errorf("expected server port 8080, got: %d", result.Server.Port)
		}
	})

	t.Run("valid HCL with arrays", func(t *testing.T) {
		data := []byte(`numbers = [1, 2, 3, 4, 5]
strings = ["a", "b", "c"]
tags = ["web", "api", "service"]`)

		type Config struct {
			Numbers []int    `hcl:"numbers"`
			Strings []string `hcl:"strings"`
			Tags    []string `hcl:"tags"`
		}
		var result Config

		err := decoder.Unmarshal(data, &result)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if len(result.Numbers) != 5 {
			t.Errorf("expected 5 numbers, got: %d", len(result.Numbers))
		}
		if result.Numbers[0] != 1 || result.Numbers[4] != 5 {
			t.Errorf("expected numbers [1,...,5], got: %v", result.Numbers)
		}

		if len(result.Strings) != 3 {
			t.Errorf("expected 3 strings, got: %d", len(result.Strings))
		}
		if result.Strings[0] != "a" || result.Strings[2] != "c" {
			t.Errorf("expected strings [a,b,c], got: %v", result.Strings)
		}

		if len(result.Tags) != 3 {
			t.Errorf("expected 3 tags, got: %d", len(result.Tags))
		}
		if result.Tags[1] != "api" {
			t.Errorf("expected second tag 'api', got: %s", result.Tags[1])
		}
	})

	t.Run("valid HCL with nested blocks", func(t *testing.T) {
		data := []byte(`app_name = "myapp"

database {
  primary {
    host = "primary.db.example.com"
    port = 5432
  }
  
  replica {
    host = "replica.db.example.com"
    port = 5432
  }
}`)

		type DBConfig struct {
			Host string `hcl:"host"`
			Port int    `hcl:"port"`
		}

		type Database struct {
			Primary DBConfig `hcl:"primary,block"`
			Replica DBConfig `hcl:"replica,block"`
		}

		type Config struct {
			AppName  string   `hcl:"app_name"`
			Database Database `hcl:"database,block"`
		}
		var result Config

		err := decoder.Unmarshal(data, &result)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if result.AppName != "myapp" {
			t.Errorf("expected app_name 'myapp', got: %s", result.AppName)
		}
		if result.Database.Primary.Host != "primary.db.example.com" {
			t.Errorf("expected primary host 'primary.db.example.com', got: %s", result.Database.Primary.Host)
		}
		if result.Database.Replica.Host != "replica.db.example.com" {
			t.Errorf("expected replica host 'replica.db.example.com', got: %s", result.Database.Replica.Host)
		}
	})

	t.Run("HCL with repeated blocks", func(t *testing.T) {
		data := []byte(`service {
  name = "web"
  port = 8080
}

service {
  name = "api"
  port = 9000
}`)

		type Service struct {
			Name string `hcl:"name"`
			Port int    `hcl:"port"`
		}

		type Config struct {
			Services []Service `hcl:"service,block"`
		}
		var result Config

		err := decoder.Unmarshal(data, &result)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if len(result.Services) != 2 {
			t.Errorf("expected 2 services, got: %d", len(result.Services))
		}
		if result.Services[0].Name != "web" {
			t.Errorf("expected first service name 'web', got: %s", result.Services[0].Name)
		}
		if result.Services[0].Port != 8080 {
			t.Errorf("expected first service port 8080, got: %d", result.Services[0].Port)
		}
		if result.Services[1].Name != "api" {
			t.Errorf("expected second service name 'api', got: %s", result.Services[1].Name)
		}
		if result.Services[1].Port != 9000 {
			t.Errorf("expected second service port 9000, got: %d", result.Services[1].Port)
		}
	})

	t.Run("empty data", func(t *testing.T) {
		data := []byte("")
		type Config struct {
			Name string `hcl:"name"`
		}
		var result Config

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

	t.Run("invalid HCL", func(t *testing.T) {
		data := []byte(`name = "test
port = 8080`)
		type Config struct {
			Name string `hcl:"name"`
			Port int    `hcl:"port"`
		}
		var result Config

		err := decoder.Unmarshal(data, &result)
		if err == nil {
			t.Fatal("expected error for invalid HCL")
		}
	})

	t.Run("HCL with simple expressions", func(t *testing.T) {
		data := []byte(`name = "test"
port = 8080
timeout = 30`)

		type Config struct {
			Name    string `hcl:"name"`
			Port    int    `hcl:"port"`
			Timeout int    `hcl:"timeout"`
		}
		var result Config

		err := decoder.Unmarshal(data, &result)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if result.Name != "test" {
			t.Errorf("expected name 'test', got: %s", result.Name)
		}
		if result.Port != 8080 {
			t.Errorf("expected port 8080, got: %d", result.Port)
		}
		if result.Timeout != 30 {
			t.Errorf("expected timeout 30, got: %d", result.Timeout)
		}
	})
}

func TestHCLDecoder_Registration(t *testing.T) {
	t.Run("format registration", func(t *testing.T) {
		dec, err := decoder.GetDecoder(FormatHCL)
		if err != nil {
			t.Fatalf("failed to get HCL decoder: %v", err)
		}
		if dec == nil {
			t.Fatal("decoder is nil")
		}
	})

	t.Run("extension mapping", func(t *testing.T) {
		testCases := []string{".hcl", ".tf"}
		for _, ext := range testCases {
			format, err := decoder.FormatFromExtension(ext)
			if err != nil {
				t.Fatalf("failed to get format from extension %s: %v", ext, err)
			}
			if format != FormatHCL {
				t.Errorf("expected format %s for extension %s, got: %s", FormatHCL, ext, format)
			}
		}
	})

	t.Run("MIME type mapping", func(t *testing.T) {
		testCases := []string{"application/hcl", "text/hcl"}
		for _, mime := range testCases {
			format, err := decoder.FormatFromMIME(mime)
			if err != nil {
				t.Fatalf("failed to get format from MIME %s: %v", mime, err)
			}
			if format != FormatHCL {
				t.Errorf("expected format %s for MIME %s, got: %s", FormatHCL, mime, format)
			}
		}
	})
}
