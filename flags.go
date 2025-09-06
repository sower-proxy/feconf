package conf

import (
	"flag"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/go-viper/mapstructure/v2"
)

// NewWithFlags creates a configuration loader that reads URI from command-line flags
// If no flag is provided or the flag value is empty, uriOrField is used
// It also parses the struct type T and adds flags for fields with 'usage' tag
func NewWithFlags[T any](uriOrField string) *ConfOpt[T] {
	// Use LoadFlags to parse struct fields and command-line flags
	flagValues, _ := LoadFlags[T]()

	// Get the actual config URI value if the flag was set
	if f := flag.Lookup(strings.ToLower(uriOrField)); f != nil {
		uriOrField = f.Value.String()
	}

	// Create a copy of DefaultParserConfig and add default values from flags
	parserConfig := DefaultParserConfig
	parserConfig.DecodeHook = mapstructure.ComposeDecodeHookFunc(
		FlagDefaultHook(flagValues),
		DefaultParserConfig.DecodeHook,
	)

	return &ConfOpt[T]{
		uri:        uriOrField,
		ParserConf: parserConfig,
	}
}

// LoadFlags parses struct fields and adds command-line flags for fields with 'usage' tag,
// parses command-line flags, maps flag values to the struct, and returns a pointer to the initialized struct
// This is a public entry function for flag parsing functionality
func LoadFlags[T any]() (*T, error) {
	// Parse struct fields and add flags for fields with usage tag
	parseStructFlags[T]()

	// Parse flags if not already parsed
	if !flag.Parsed() {
		flag.Parse()
	}

	// Create a zero value of T
	var result T

	// Create a map to store flag values
	flagValues := make(map[string]any)

	// Visit all set flags and collect their values
	flag.Visit(func(f *flag.Flag) {
		flagValues[f.Name] = f.Value.String()
	})

	// Use mapstructure to decode flag values to struct
	decoderConfig := &mapstructure.DecoderConfig{
		Result:           &result,
		WeaklyTypedInput: true,
		TagName:          "default",
	}

	mapDecoder, err := mapstructure.NewDecoder(decoderConfig)
	if err != nil {
		// If decoder creation fails, return error
		return nil, fmt.Errorf("failed to create mapstructure decoder: %w", err)
	}

	// Decode flag values to struct
	if err := mapDecoder.Decode(flagValues); err != nil {
		// If decoding fails, return error
		return nil, fmt.Errorf("failed to decode flag values to struct: %w", err)
	}

	return &result, nil
}

// FlagDefaultHook creates a decode hook that provides default values from flag-parsed struct
func FlagDefaultHook(flagValues any) mapstructure.DecodeHookFuncType {
	return func(f reflect.Type, t reflect.Type, data any) (any, error) {
		// If data is zero value, try to get default value from flagValues
		if isZeroValue(data) {
			if defaultVal := getFlagDefaultValue(flagValues, t); defaultVal != nil {
				return defaultVal, nil
			}
		}

		// If data is not zero value, or no default found, return original data
		return data, nil
	}
}

// getFlagDefaultValue extracts default value from flagValues struct for a given type
func getFlagDefaultValue(flagValues any, targetType reflect.Type) any {
	if flagValues == nil {
		return nil
	}

	val := reflect.ValueOf(flagValues)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return nil
	}

	// Try to find a field in flagValues that matches the target type
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := val.Type().Field(i)

		// Skip unexported fields
		if fieldType.PkgPath != "" {
			continue
		}

		// Check if field type matches target type
		if field.Type() == targetType {
			return field.Interface()
		}
	}

	return nil
}

// parseStructFlags parses struct fields and adds flags for fields with usage tag
func parseStructFlags[T any]() {
	var zero T
	typ := reflect.TypeOf(zero)

	// If it's a pointer, get the underlying type
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	// Only process struct types
	if typ.Kind() != reflect.Struct {
		return
	}

	// Iterate through struct fields
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		// Skip unexported fields
		if field.PkgPath != "" {
			continue
		}

		// Get field tags
		usage := field.Tag.Get("usage")
		defaultValue := field.Tag.Get("default")

		// Only add flag if usage tag is present
		if usage != "" {
			flagName := strings.ToLower(field.Name)

			// Check if flag already exists
			var flagExists bool
			flag.VisitAll(func(f *flag.Flag) {
				if f.Name == flagName {
					flagExists = true
				}
			})

			if !flagExists {
				// Add flag based on field type
				switch field.Type.Kind() {
				case reflect.String:
					flag.String(flagName, defaultValue, usage)
				case reflect.Bool:
					defaultBool := false
					if defaultValue != "" {
						defaultBool = strings.ToLower(defaultValue) == "true"
					}
					flag.Bool(flagName, defaultBool, usage)
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					defaultInt := 0
					if defaultValue != "" {
						if val, err := strconv.Atoi(defaultValue); err == nil {
							defaultInt = val
						}
					}
					flag.Int(flagName, defaultInt, usage)
				case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
					defaultUint := uint(0)
					if defaultValue != "" {
						if val, err := strconv.ParseUint(defaultValue, 10, 64); err == nil {
							defaultUint = uint(val)
						}
					}
					flag.Uint(flagName, defaultUint, usage)
				case reflect.Float32, reflect.Float64:
					defaultFloat := 0.0
					if defaultValue != "" {
						if val, err := strconv.ParseFloat(defaultValue, 64); err == nil {
							defaultFloat = val
						}
					}
					flag.Float64(flagName, defaultFloat, usage)
				case reflect.Slice:
					if field.Type.Elem().Kind() == reflect.String {
						// For string slices, use String (comma-separated)
						flag.String(flagName, defaultValue, usage)
					}
				}
			}
		}
	}
}
