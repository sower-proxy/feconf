package feconf

import (
	"context"
	"flag"
	"os"
	"strconv"
	"testing"
	"time"
)

func TestNewWithFlags_StructFlags(t *testing.T) {
	// Reset flag set for clean test
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Save original args and restore them after test
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Set test args using the actual flag names that will be created
	os.Args = []string{"test", "--host", "testhost", "--port", "9090", "--debug", "true"}

	// Create configuration with struct flags using a URI to avoid flag name conflict
	confOpt := NewWithFlags[TestConfig]("file:///tmp/config.json")
	defer confOpt.Close()

	// Load flags to parse struct fields and command line arguments
	var config TestConfig
	err := LoadFlags(&config)
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
		t.Errorf("Expected host flag with correct usage, got: %+v", hostFlag)
	}

	if portFlag == nil || portFlag.Usage != "Server port" {
		t.Errorf("Expected port flag with correct usage, got: %+v", portFlag)
	}

	if debugFlag == nil || debugFlag.Usage != "Enable debug mode" {
		t.Errorf("Expected debug flag with correct usage, got: %+v", debugFlag)
	}

	if featuresFlag == nil || featuresFlag.Usage != "Enabled features" {
		t.Errorf("Expected features flag with correct usage, got: %+v", featuresFlag)
	}

	if timeoutFlag == nil || timeoutFlag.Usage != "Request timeout in seconds" {
		t.Errorf("Expected timeout flag with correct usage, got: %+v", timeoutFlag)
	}

	// Verify flag default values
	if hostFlag != nil && hostFlag.DefValue != "localhost" {
		t.Errorf("Expected host flag default value 'localhost', got '%s'", hostFlag.DefValue)
	}
	if portFlag != nil && portFlag.DefValue != "8080" {
		t.Errorf("Expected port flag default value '8080', got '%s'", portFlag.DefValue)
	}
	if debugFlag != nil && debugFlag.DefValue != "false" {
		t.Errorf("Expected debug flag default value 'false', got '%s'", debugFlag.DefValue)
	}
	if featuresFlag != nil && featuresFlag.DefValue != "feature1,feature2" {
		t.Errorf("Expected features flag default value 'feature1,feature2', got '%s'", featuresFlag.DefValue)
	}
	if timeoutFlag != nil && timeoutFlag.DefValue != "30" {
		t.Errorf("Expected timeout flag default value '30', got '%s'", timeoutFlag.DefValue)
	}

	// Note: LoadFlags only maps flags that were explicitly set via command line
	// The config struct values are not directly set from command line flags in this implementation
	// This test focuses on verifying that struct flags are created correctly

	// Verify that the flags were set correctly from command line
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

	// Create configuration with struct flags using URI to avoid conflicts
	confOpt := NewWithFlags[TestConfig]("file:///tmp/config.json")
	defer confOpt.Close()

	// Load flags to parse struct fields
	var flagValues TestConfig
	err := LoadFlags(&flagValues)
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

	// Verify default values in flag definitions
	if hostFlag == nil || hostFlag.DefValue != "localhost" {
		t.Error("Expected host flag with default value 'localhost'")
	}

	if portFlag == nil || portFlag.DefValue != "8080" {
		t.Error("Expected port flag with default value '8080'")
	}

	if debugFlag == nil || debugFlag.DefValue != "false" {
		t.Error("Expected debug flag with default value 'false'")
	}

	// When no flags are set via command line, LoadFlags should not set any values
	// since flag.Visit only collects flags that were explicitly set
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

	// Load flags
	var config TestConfig
	err := LoadFlags(&config)
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

func TestNewWithFlags_PointerType(t *testing.T) {
	// Reset flag set for clean test
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Save original args and restore them after test
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Set test args
	os.Args = []string{"test"}

	// Create configuration with pointer type using URI to avoid conflicts
	confOpt := NewWithFlags[*SimpleConfig]("file:///tmp/config.json")
	defer confOpt.Close()

	// Load flags to parse struct fields
	var flagValues *SimpleConfig
	err := LoadFlags(&flagValues)
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

	// When no flags are set, pointer should be nil or zero values
	if flagValues != nil && (*flagValues).Name != "" {
		t.Errorf("Expected name flag to be empty when no flag set, got '%s'", (*flagValues).Name)
	}
}

func TestLoadFlagsCtx_ContextCancellation(t *testing.T) {
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

	// Load flags with cancelled context
	var config TestConfig
	err := LoadFlagsCtx(ctx, &config)
	if err != context.Canceled {
		t.Errorf("Expected context.Canceled error, got %v", err)
	}
}

func TestParseStructFlagsCtx_ContextCancellation(t *testing.T) {
	// Reset flag set for clean test
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// This should not panic and should return early due to cancelled context
	parseStructFlagsCtx[TestConfig](ctx)

	// No flags should have been created
	flagCount := 0
	flag.VisitAll(func(f *flag.Flag) {
		flagCount++
	})

	if flagCount != 0 {
		t.Errorf("Expected no flags to be created with cancelled context, got %d", flagCount)
	}
}

func TestParseStructFlagsCtx_Timeout(t *testing.T) {
	// Reset flag set for clean test
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Create context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	// Wait for context to timeout
	time.Sleep(20 * time.Millisecond)

	// This should return early due to timeout
	parseStructFlagsCtx[TestConfig](ctx)

	// No flags should have been created
	flagCount := 0
	flag.VisitAll(func(f *flag.Flag) {
		flagCount++
	})

	if flagCount != 0 {
		t.Errorf("Expected no flags to be created with timed out context, got %d", flagCount)
	}
}

func TestLoadWithFlags_ContextCancellation(t *testing.T) {
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

	// Test LoadWithFlagsCtx with cancelled context
	var config TestConfig
	err := LoadWithFlagsCtx(ctx, &config, "file:///tmp/config.json")
	if err != context.Canceled {
		t.Errorf("Expected context.Canceled error, got %v", err)
	}
}
