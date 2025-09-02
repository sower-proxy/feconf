package conf

import (
	"flag"
	"os"
	"testing"
)

func TestNewWithFlags(t *testing.T) {
	// Save original command line args
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Reset flag package for testing
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	tests := []struct {
		name        string
		args        []string
		defaultURI  string
		expectedURI string
	}{
		{
			name:        "no flag provided - use default",
			args:        []string{"program"},
			defaultURI:  "file://./config.json",
			expectedURI: "file://./config.json",
		},
		{
			name:        "flag provided with value",
			args:        []string{"program", "-config", "redis://localhost:6379/app-config"},
			defaultURI:  "file://./config.json",
			expectedURI: "redis://localhost:6379/app-config",
		},
		{
			name:        "flag provided with empty value - use default",
			args:        []string{"program", "-config", ""},
			defaultURI:  "file://./config.json",
			expectedURI: "file://./config.json",
		},
		{
			name:        "flag provided with complex URI",
			args:        []string{"program", "-config", "redis://localhost:6379/settings?db=1&content-type=application/json#database"},
			defaultURI:  "file://./config.json",
			expectedURI: "redis://localhost:6379/settings?db=1&content-type=application/json#database",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset flag package for each test
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
			
			// Set command line args
			os.Args = tt.args

			// Create config loader
			loader := NewWithFlags[TestConfig](tt.defaultURI)
			defer loader.Close()

			// Check if URI was set correctly
			if loader.uri != tt.expectedURI {
				t.Errorf("NewWithFlags() uri = %v, want %v", loader.uri, tt.expectedURI)
			}
		})
	}
}

func TestNewWithFlagsNamed(t *testing.T) {
	// Save original command line args
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	tests := []struct {
		name        string
		flagName    string
		args        []string
		defaultURI  string
		expectedURI string
	}{
		{
			name:        "custom flag name - no flag provided",
			flagName:    "app-config",
			args:        []string{"program"},
			defaultURI:  "file://./app.yaml",
			expectedURI: "file://./app.yaml",
		},
		{
			name:        "custom flag name - flag provided",
			flagName:    "app-config",
			args:        []string{"program", "-app-config", "http://config-server/app.json"},
			defaultURI:  "file://./app.yaml",
			expectedURI: "http://config-server/app.json",
		},
		{
			name:        "custom flag name with equals format",
			flagName:    "settings",
			args:        []string{"program", "-settings=redis://localhost:6379/settings"},
			defaultURI:  "file://./settings.toml",
			expectedURI: "redis://localhost:6379/settings",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset flag package for each test
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
			
			// Set command line args
			os.Args = tt.args

			// Create config loader with custom flag name
			loader := NewWithFlagsNamed[TestConfig](tt.flagName, tt.defaultURI)
			defer loader.Close()

			// Check if URI was set correctly
			if loader.uri != tt.expectedURI {
				t.Errorf("NewWithFlagsNamed() uri = %v, want %v", loader.uri, tt.expectedURI)
			}
		})
	}
}

func TestGetFlagValue(t *testing.T) {
	// Save original command line args
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	tests := []struct {
		name         string
		flagName     string
		args         []string
		defaultValue string
		expected     string
	}{
		{
			name:         "flag not provided",
			flagName:     "test-flag",
			args:         []string{"program"},
			defaultValue: "default-value",
			expected:     "default-value",
		},
		{
			name:         "flag provided with value",
			flagName:     "test-flag",
			args:         []string{"program", "-test-flag", "provided-value"},
			defaultValue: "default-value",
			expected:     "provided-value",
		},
		{
			name:         "flag provided with empty value",
			flagName:     "test-flag",
			args:         []string{"program", "-test-flag", ""},
			defaultValue: "default-value",
			expected:     "default-value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset flag package for each test
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
			
			// Set command line args
			os.Args = tt.args

			// Test getFlagValue function
			result := getFlagValue(tt.flagName, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("getFlagValue() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestConfig is a test configuration structure
type TestConfig struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

func TestIntegrationWithFlags(t *testing.T) {
	// This test requires actual config files, skip if not available
	t.Skip("Integration test - requires actual config files")

	// Save original command line args
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Reset flag package
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Simulate command line with config flag
	os.Args = []string{"program", "-config", "file://./examples/file-json/config.json"}

	// Create loader
	loader := NewWithFlags[TestConfig]("file://./default-config.json")
	defer loader.Close()

	// Try to load configuration
	config, err := loader.Load()
	if err != nil {
		t.Logf("Expected error for integration test: %v", err)
		return
	}

	// Verify config was loaded
	if config == nil {
		t.Error("Config should not be nil")
	}
}