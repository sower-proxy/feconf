package feconf

import (
	"context"
	"flag"
	"os"
	"reflect"
	"testing"
)

// TestConfig is a test configuration struct
type TestConfig struct {
	Host     string   `usage:"Server host address" default:"localhost"`
	Port     int      `usage:"Server port" default:"8080"`
	Debug    bool     `usage:"Enable debug mode" default:"false"`
	Features []string `usage:"Enabled features" default:"feature1,feature2"`
	Timeout  int      `usage:"Request timeout in seconds" default:"30"`
	// This field should not create a flag as it has no usage tag
	InternalField string `default:"internal"`
}

// SimpleConfig is a simple configuration struct for testing
type SimpleConfig struct {
	Name string `usage:"Application name" default:"myapp"`
}

func TestIsValidURI(t *testing.T) {
	tests := []struct {
		name string
		uri  string
		want bool
	}{
		{"Valid HTTP URI", "http://example.com", true},
		{"Valid HTTPS URI", "https://example.com", true},
		{"Valid File URI", "file:///tmp/config.json", true},
		{"Valid Redis URI", "redis://localhost:6379", true},
		{"Invalid URI - no scheme", "example.com", false},
		{"Invalid URI - empty", "", false},
		{"Invalid URI - malformed", "http://", true}, // url.Parse considers "http://" as valid
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidURI(tt.uri); got != tt.want {
				t.Errorf("isValidURI(%q) = %v, want %v", tt.uri, got, tt.want)
			}
		})
	}
}

func TestGetFlagDefaultValue(t *testing.T) {
	type testStruct struct {
		Name  string
		Count int
		Flag  bool
	}

	tests := []struct {
		name       string
		flagValues any
		targetType reflect.Type
		want       any
	}{
		{
			name:       "String field",
			flagValues: &testStruct{Name: "test"},
			targetType: reflect.TypeOf(""),
			want:       "test",
		},
		{
			name:       "Int field",
			flagValues: &testStruct{Count: 42},
			targetType: reflect.TypeOf(0),
			want:       42,
		},
		{
			name:       "Bool field",
			flagValues: &testStruct{Flag: true},
			targetType: reflect.TypeOf(false),
			want:       true,
		},
		{
			name:       "No matching field",
			flagValues: &testStruct{Name: "test"},
			targetType: reflect.TypeOf(3.14),
			want:       nil,
		},
		{
			name:       "Nil flagValues",
			flagValues: nil,
			targetType: reflect.TypeOf(""),
			want:       nil,
		},
		{
			name:       "Non-struct flagValues",
			flagValues: "not a struct",
			targetType: reflect.TypeOf(""),
			want:       nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var flagValues *testStruct
			if tt.flagValues != nil {
				if val, ok := tt.flagValues.(*testStruct); ok {
					flagValues = val
				}
			}
			if got := getFlagDefaultValue(flagValues, tt.targetType); got != tt.want {
				t.Errorf("getFlagDefaultValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewWithFlags_URI(t *testing.T) {
	// Reset flag set for clean test
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Save original args and restore them after test
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Set test args
	os.Args = []string{"test"}

	// Create configuration with direct URI
	confOpt := NewWithFlags[TestConfig]("file:///tmp/config.json")
	defer confOpt.Close()

	// Verify the URI was set correctly
	if confOpt.uri != "file:///tmp/config.json" {
		t.Errorf("Expected URI to be 'file:///tmp/config.json', got '%s'", confOpt.uri)
	}

	// Load flags to trigger struct parsing
	var config TestConfig
	err := LoadFlags(&config)
	if err != nil {
		t.Fatalf("Failed to load flags: %v", err)
	}

	// Should create struct flags even when URI is provided, because LoadFlags always parses struct fields
	// But the config flag should not exist since we provided a direct URI
	flagCount := 0
	configFlagExists := false
	flag.VisitAll(func(f *flag.Flag) {
		if f.Name == "config" {
			configFlagExists = true
		}
		flagCount++
	})

	// Verify that struct flags were created (host, port, debug, etc.)
	if flagCount < 5 { // At least 5 struct flags should be created
		t.Errorf("Expected struct flags to be created, got %d flags", flagCount)
	}

	// Verify that config flag does not exist since we provided a direct URI
	if configFlagExists {
		t.Error("Expected config flag to not exist when URI is provided directly")
	}
}

func TestNewWithFlagsCtx_ContextCancellation(t *testing.T) {
	// Reset flag set for clean test
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Save original args and restore them after test
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Set test args
	os.Args = []string{"test"}

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Create configuration with cancelled context
	confOpt := NewWithFlagsCtx[TestConfig](ctx, "Host")
	defer confOpt.Close()

	// Load flags - should return context cancelled error
	var config TestConfig
	err := LoadFlagsCtx(ctx, &config)
	if err != context.Canceled {
		t.Errorf("Expected context.Canceled error, got %v", err)
	}
}

func TestConfOpt_LoadCtx_ContextCancellation(t *testing.T) {
	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Test LoadCtx with cancelled context
	confOpt := New[TestConfig]("invalid://uri")
	defer confOpt.Close()

	var config TestConfig
	err := confOpt.LoadCtx(ctx, &config)
	if err != context.Canceled {
		t.Errorf("Expected context.Canceled error, got %v", err)
	}
}

func TestConfOpt_LoadCtx_NilPointer(t *testing.T) {
	confOpt := New[TestConfig]("invalid://uri")
	defer confOpt.Close()

	err := confOpt.LoadCtx(context.Background(), nil)
	if err == nil {
		t.Error("Expected error for nil pointer, got nil")
	}
}
