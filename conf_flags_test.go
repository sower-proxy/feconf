package conf

import (
	"flag"
	"os"
	"strconv"
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

func TestNewWithFlags_StructFlags(t *testing.T) {
	// Reset flag set for clean test
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Save original args and restore them after test
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Set test args
	os.Args = []string{"test", "--host", "testhost", "--port", "9090", "--debug", "true"}

	// Create configuration with struct flags using field name as parameter
	confOpt := NewWithFlags[TestConfig]("Host")
	defer confOpt.Close()

	// Load flags first to get default values
	_, err := LoadFlags[TestConfig]()
	if err != nil {
		t.Fatalf("Failed to load flags: %v", err)
	}

	// Check if flags were properly defined
	var hostFlag, portFlag, debugFlag, featuresFlag, timeoutFlag *flag.Flag
	flag.VisitAll(func(f *flag.Flag) {
		switch f.Name {
		case "host":
			hostFlag = f
		case "port":
			portFlag = f
		case "debug":
			debugFlag = f
		case "features":
			featuresFlag = f
		case "timeout":
			timeoutFlag = f
		}
	})

	// Verify flags were created with correct usage
	if hostFlag == nil || hostFlag.Usage != "Server host address" {
		t.Error("Expected host flag with correct usage")
	}

	if portFlag == nil || portFlag.Usage != "Server port" {
		t.Error("Expected port flag with correct usage")
	}

	if debugFlag == nil || debugFlag.Usage != "Enable debug mode" {
		t.Error("Expected debug flag with correct usage")
	}

	if featuresFlag == nil || featuresFlag.Usage != "Enabled features" {
		t.Error("Expected features flag with correct usage")
	}

	if timeoutFlag == nil || timeoutFlag.Usage != "Request timeout in seconds" {
		t.Error("Expected timeout flag with correct usage")
	}

	// Verify flag values were loaded correctly by checking actual flag values
	var actualHost, actualPort, actualDebug any
	flag.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "host":
			actualHost = f.Value.String()
		case "port":
			if val, err := strconv.Atoi(f.Value.String()); err == nil {
				actualPort = val
			}
		case "debug":
			if val, err := strconv.ParseBool(f.Value.String()); err == nil {
				actualDebug = val
			}
		}
	})

	if actualHost != "testhost" {
		t.Errorf("Expected host flag value 'testhost', got '%v'", actualHost)
	}
	if actualPort != 9090 {
		t.Errorf("Expected port flag value 9090, got %v", actualPort)
	}
	if actualDebug != true {
		t.Errorf("Expected debug flag value true, got %v", actualDebug)
	}
}

func TestNewWithFlags_DefaultValues(t *testing.T) {
	// Reset flag set for clean test
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Save original args and restore them after test
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Set test args with no custom flags
	os.Args = []string{"test"}

	// Create configuration with struct flags
	confOpt := NewWithFlags[TestConfig]("Host")
	defer confOpt.Close()

	// Load flags first to get default values
	flagValues, err := LoadFlags[TestConfig]()
	if err != nil {
		t.Fatalf("Failed to load flags: %v", err)
	}

	// Check if flags were properly defined with default values
	var hostFlag, portFlag, debugFlag *flag.Flag
	flag.VisitAll(func(f *flag.Flag) {
		switch f.Name {
		case "host":
			hostFlag = f
		case "port":
			portFlag = f
		case "debug":
			debugFlag = f
		}
	})

	// Verify default values
	if hostFlag == nil || hostFlag.DefValue != "localhost" {
		t.Error("Expected host flag with default value 'localhost'")
	}

	if portFlag == nil || portFlag.DefValue != "8080" {
		t.Error("Expected port flag with default value '8080'")
	}

	if debugFlag == nil || debugFlag.DefValue != "false" {
		t.Error("Expected debug flag with default value 'false'")
	}

	// Verify flag values were loaded with defaults
	// Note: When no flags are set, flagValues will have zero values
	// The actual default values are handled by the flag system itself
	if flagValues.Host != "" {
		t.Errorf("Expected host flag to be empty when no flag set, got '%s'", flagValues.Host)
	}
	if flagValues.Port != 0 {
		t.Errorf("Expected port flag to be 0 when no flag set, got %d", flagValues.Port)
	}
	if flagValues.Debug != false {
		t.Errorf("Expected debug flag to be false when no flag set, got %t", flagValues.Debug)
	}
}

func TestNewWithFlags_NoUsageTag(t *testing.T) {
	// Reset flag set for clean test
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Save original args and restore them after test
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Set test args
	os.Args = []string{"test"}

	// Create configuration with struct flags
	confOpt := NewWithFlags[TestConfig]("Host")
	defer confOpt.Close()

	// Load flags first
	_, err := LoadFlags[TestConfig]()
	if err != nil {
		t.Fatalf("Failed to load flags: %v", err)
	}

	// Check that internalfield flag was not created
	var internalFieldFlag *flag.Flag
	flag.VisitAll(func(f *flag.Flag) {
		if f.Name == "internalfield" {
			internalFieldFlag = f
		}
	})

	// Verify that internal field flag was not created
	if internalFieldFlag != nil {
		t.Error("Expected internalfield flag to not be created (no usage tag)")
	}
}

// SimpleConfig is a simple configuration struct for testing
type SimpleConfig struct {
	Name string `usage:"Application name" default:"myapp"`
}

func TestNewWithFlags_PointerType(t *testing.T) {
	// Reset flag set for clean test
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Save original args and restore them after test
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Set test args
	os.Args = []string{"test"}

	// Create configuration with pointer type
	confOpt := NewWithFlags[*SimpleConfig]("Name")
	defer confOpt.Close()

	// Load flags first
	flagValues, err := LoadFlags[*SimpleConfig]()
	if err != nil {
		t.Fatalf("Failed to load flags: %v", err)
	}

	// Check if flag was properly defined
	var nameFlag *flag.Flag
	flag.VisitAll(func(f *flag.Flag) {
		if f.Name == "name" {
			nameFlag = f
		}
	})

	// Verify flag was created
	if nameFlag == nil || nameFlag.Usage != "Application name" {
		t.Error("Expected name flag with correct usage")
	}

	if nameFlag == nil || nameFlag.DefValue != "myapp" {
		t.Error("Expected name flag with default value 'myapp'")
	}

	// Verify flag values were loaded correctly
	// Note: When no flags are set, flagValues will have zero values
	if flagValues != nil && (*flagValues).Name != "" {
		t.Errorf("Expected name flag to be empty when no flag set, got '%s'", (*flagValues).Name)
	}
}
