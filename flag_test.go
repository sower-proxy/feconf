package feconf

import (
	"flag"
	"os"
	"reflect"
	"testing"
)

// resetFlags resets the flag package state for testing
// This is a workaround since flag package uses global state
func resetFlags() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	globalFlags = make(map[any]map[string]*flagValue)
}

func TestRegisterFlagsFromStruct(t *testing.T) {
	tests := []struct {
		name           string
		structType     any
		expectedFlags  []string
		unexpectedFlag string
	}{
		{
			name: "struct with usage tags",
			structType: struct {
				Server string `json:"server" usage:"server address"`
				Port   int    `json:"port" usage:"server port"`
			}{},
			expectedFlags: []string{"server", "port"},
		},
		{
			name: "struct without usage tags",
			structType: struct {
				Server string `json:"server"`
				Port   int    `json:"port"`
			}{},
			expectedFlags:  []string{},
			unexpectedFlag: "server",
		},
		{
			name: "struct with mixed tags",
			structType: struct {
				Server  string `json:"server" usage:"server address"`
				Port    int    `json:"port"`
				Enabled bool   `json:"enabled" usage:"enable feature"`
			}{},
			expectedFlags:  []string{"server", "enabled"},
			unexpectedFlag: "port",
		},
		{
			name: "struct using json tag",
			structType: struct {
				Server string `json:"server" usage:"server address"`
			}{},
			expectedFlags: []string{"server"},
		},
		{
			name: "struct using field name",
			structType: struct {
				Server string `usage:"server address"`
			}{},
			expectedFlags: []string{"Server"},
		},
		{
			name: "struct with omitempty",
			structType: struct {
				Server string `json:"server,omitempty" usage:"server address"`
			}{},
			expectedFlags: []string{"server"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetFlags()

			conf := &ConfOpt[struct{}]{
				uri: "test.json",
			}
			globalFlags[conf] = make(map[string]*flagValue)
			conf.registerStructFlags(reflect.TypeOf(tt.structType))

			flags := globalFlags[conf]

			for _, expectedFlag := range tt.expectedFlags {
				if _, ok := flags[expectedFlag]; !ok {
					t.Errorf("Expected flag %s to be registered", expectedFlag)
				}
			}

			if tt.unexpectedFlag != "" {
				if _, ok := flags[tt.unexpectedFlag]; ok {
					t.Errorf("Unexpected flag %s should not be registered", tt.unexpectedFlag)
				}
			}
		})
	}
}

func TestRegisterFlagsFromStruct_Types(t *testing.T) {
	resetFlags()

	type TestConfig struct {
		StringVal  string  `json:"string_val" usage:"string value"`
		IntVal     int     `json:"int_val" usage:"int value"`
		Int64Val   int64   `json:"int64_val" usage:"int64 value"`
		UintVal    uint    `json:"uint_val" usage:"uint value"`
		BoolVal    bool    `json:"bool_val" usage:"bool value"`
		Float64Val float64 `json:"float64_val" usage:"float64 value"`
	}

	conf := &ConfOpt[TestConfig]{
		uri: "test.json",
	}
	globalFlags[conf] = make(map[string]*flagValue)
	conf.registerStructFlags(reflect.TypeOf(TestConfig{}))

	flags := globalFlags[conf]

	expectedFlags := []string{"string_val", "int_val", "int64_val", "uint_val", "bool_val", "float64_val"}
	for _, expectedFlag := range expectedFlags {
		if _, ok := flags[expectedFlag]; !ok {
			t.Errorf("Expected flag %s to be registered", expectedFlag)
		}
	}
}

func TestRegisterFlagsFromStruct_PointerType(t *testing.T) {
	resetFlags()

	type TestConfig struct {
		Server string `json:"server" usage:"server address"`
	}

	conf := &ConfOpt[TestConfig]{
		uri: "test.json",
	}
	globalFlags[conf] = make(map[string]*flagValue)

	// Test with pointer type
	conf.registerStructFlags(reflect.TypeOf(&TestConfig{}))

	flags := globalFlags[conf]
	if _, ok := flags["server"]; !ok {
		t.Error("Expected flag 'server' to be registered for pointer type")
	}
}

func TestRegisterFlagsFromStruct_NilType(t *testing.T) {
	resetFlags()

	conf := &ConfOpt[struct{}]{
		uri: "test.json",
	}
	globalFlags[conf] = make(map[string]*flagValue)

	// Should not panic with nil type
	conf.registerStructFlags(nil)

	if len(globalFlags[conf]) != 0 {
		t.Error("Expected no flags to be registered for nil type")
	}
}

func TestRegisterFlagsFromStruct_NonStructType(t *testing.T) {
	resetFlags()

	conf := &ConfOpt[string]{
		uri: "test.json",
	}
	globalFlags[conf] = make(map[string]*flagValue)

	// Should not panic with non-struct type
	conf.registerStructFlags(reflect.TypeOf("string"))

	if len(globalFlags[conf]) != 0 {
		t.Error("Expected no flags to be registered for non-struct type")
	}
}

func TestRegisterFlags_WithFlagName(t *testing.T) {
	resetFlags()

	type TestConfig struct {
		Server string `json:"server" usage:"server address"`
	}

	conf := &ConfOpt[TestConfig]{
		uri:        "",
		flagName:   "config",
		ParserConf: DefaultParserConfig,
	}
	conf.registerFlags()

	// Check that -config flag is registered when flagName is set
	configFlag := flag.Lookup("config")
	if configFlag == nil {
		t.Error("Expected -config flag to be registered when flagName is set")
	}

	// Check that struct field flags are also registered
	flags := globalFlags[conf]
	if _, ok := flags["server"]; !ok {
		t.Error("Expected flag 'server' to be registered")
	}
}

func TestRegisterFlags_WithoutFlagName(t *testing.T) {
	resetFlags()

	type TestConfig struct {
		Server string `json:"server" usage:"server address"`
	}

	conf := &ConfOpt[TestConfig]{
		uri:        "config.json",
		flagName:   "",
		ParserConf: DefaultParserConfig,
	}
	conf.registerFlags()

	// Check that -config flag is NOT registered when flagName is empty
	configFlag := flag.Lookup("config")
	if configFlag != nil {
		t.Error("Expected -config flag to NOT be registered when flagName is empty")
	}

	// Check that struct field flags are still registered
	flags := globalFlags[conf]
	if _, ok := flags["server"]; !ok {
		t.Error("Expected flag 'server' to be registered")
	}
}

func TestMergeFlagValues(t *testing.T) {
	resetFlags()

	type TestConfig struct {
		Server string `json:"server" usage:"server address"`
		Port   int    `json:"port" usage:"server port"`
	}

	conf := &ConfOpt[TestConfig]{
		uri:        "config.json",
		ParserConf: DefaultParserConfig,
		parsedData: map[string]any{
			"server": "localhost",
			"port":   8080,
		},
	}
	globalFlags[conf] = make(map[string]*flagValue)

	// Simulate flag values being set
	serverVal := "override-server"
	portVal := 9090
	globalFlags[conf]["server"] = &flagValue{ptr: &serverVal, isSet: true}
	globalFlags[conf]["port"] = &flagValue{ptr: &portVal, isSet: false} // Not set by user

	conf.mergeFlagValues()

	// Server should be overridden
	if conf.parsedData["server"] != "override-server" {
		t.Errorf("Expected server to be 'override-server', got %v", conf.parsedData["server"])
	}

	// Port should remain unchanged (flag not set)
	if conf.parsedData["port"] != 8080 {
		t.Errorf("Expected port to remain 8080, got %v", conf.parsedData["port"])
	}
}

func TestMergeFlagValues_AllTypes(t *testing.T) {
	resetFlags()

	conf := &ConfOpt[struct{}]{
		uri:        "config.json",
		ParserConf: DefaultParserConfig,
		parsedData: map[string]any{},
	}
	globalFlags[conf] = make(map[string]*flagValue)

	// Set up flags of different types
	stringVal := "test-string"
	intVal := 42
	uintVal := uint(100)
	boolVal := true
	floatVal := 3.14

	globalFlags[conf]["string"] = &flagValue{ptr: &stringVal, isSet: true}
	globalFlags[conf]["int"] = &flagValue{ptr: &intVal, isSet: true}
	globalFlags[conf]["uint"] = &flagValue{ptr: &uintVal, isSet: true}
	globalFlags[conf]["bool"] = &flagValue{ptr: &boolVal, isSet: true}
	globalFlags[conf]["float"] = &flagValue{ptr: &floatVal, isSet: true}

	conf.mergeFlagValues()

	if conf.parsedData["string"] != "test-string" {
		t.Errorf("Expected string to be 'test-string', got %v", conf.parsedData["string"])
	}
	if conf.parsedData["int"] != 42 {
		t.Errorf("Expected int to be 42, got %v", conf.parsedData["int"])
	}
	if conf.parsedData["uint"] != uint(100) {
		t.Errorf("Expected uint to be 100, got %v", conf.parsedData["uint"])
	}
	if conf.parsedData["bool"] != true {
		t.Errorf("Expected bool to be true, got %v", conf.parsedData["bool"])
	}
	if conf.parsedData["float"] != 3.14 {
		t.Errorf("Expected float to be 3.14, got %v", conf.parsedData["float"])
	}
}

func TestMergeFlagValues_NilParsedData(t *testing.T) {
	resetFlags()

	conf := &ConfOpt[struct{}]{
		uri:        "config.json",
		ParserConf: DefaultParserConfig,
		parsedData: nil,
	}
	globalFlags[conf] = make(map[string]*flagValue)

	stringVal := "test"
	globalFlags[conf]["test"] = &flagValue{ptr: &stringVal, isSet: true}

	conf.mergeFlagValues()

	// Should create parsedData and add the value
	if conf.parsedData == nil {
		t.Error("Expected parsedData to be initialized")
	}
	if conf.parsedData["test"] != "test" {
		t.Errorf("Expected test to be 'test', got %v", conf.parsedData["test"])
	}
}

func TestMergeFlagValues_NilFlags(t *testing.T) {
	resetFlags()

	conf := &ConfOpt[struct{}]{
		uri:        "config.json",
		ParserConf: DefaultParserConfig,
		parsedData: map[string]any{"existing": "value"},
	}
	// Don't initialize globalFlags[conf]

	// Should not panic
	conf.mergeFlagValues()

	// Existing data should remain
	if conf.parsedData["existing"] != "value" {
		t.Errorf("Expected existing value to remain, got %v", conf.parsedData["existing"])
	}
}

func TestParseFlags(t *testing.T) {
	resetFlags()

	type TestConfig struct {
		Server string `json:"server" usage:"server address"`
	}

	conf := &ConfOpt[TestConfig]{
		uri:        "config.json",
		ParserConf: DefaultParserConfig,
	}
	conf.registerFlags()

	// Simulate command line arguments
	os.Args = []string{"test", "-server=test-server"}
	flag.CommandLine.Parse(os.Args[1:])

	conf.parseFlags()

	flags := globalFlags[conf]
	if fv, ok := flags["server"]; ok {
		if !fv.isSet {
			t.Error("Expected 'server' flag to be marked as set")
		}
		if *(fv.ptr.(*string)) != "test-server" {
			t.Errorf("Expected server value to be 'test-server', got %v", *(fv.ptr.(*string)))
		}
	} else {
		t.Error("Expected 'server' flag to exist")
	}
}
