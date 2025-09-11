package feconf

import (
	"context"
	"flag"
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"github.com/go-viper/mapstructure/v2"
)

// NewWithFlags creates a configuration loader that reads URI from command-line flags
// If uriOrFlag is a valid URI, it will be used as the configuration source
// If uriOrFlag is not a URI, it will be treated as a flag name for configuration
// It also parses the struct type T and adds flags for fields with 'usage' tag
func NewWithFlags[T any](uriOrFlag string) *ConfOpt[T] {
	return NewWithFlagsCtx[T](context.Background(), uriOrFlag)
}

// NewWithFlagsCtx creates a configuration loader that reads URI from command-line flags with context support
// If uriOrFlag is a valid URI, it will be used as the configuration source
// If uriOrFlag is not a URI, it will be treated as a flag name for configuration
// It also parses the struct type T and adds flags for fields with 'usage' tag
// The context can be used for cancellation or timeout during flag parsing
func NewWithFlagsCtx[T any](ctx context.Context, uriOrFlag string) *ConfOpt[T] {
	isURI := isValidURI(uriOrFlag)
	if !isURI && flag.Lookup(strings.ToLower(uriOrFlag)) == nil {
		flag.String(strings.ToLower(uriOrFlag), "", "Configuration URI")
	}

	// Use LoadFlagsCtx to parse struct fields and command-line flags with context
	var flagValues T
	_ = LoadFlagsCtx(ctx, &flagValues)

	if !isURI {
		// Get the flag value after parsing
		if f := flag.Lookup(strings.ToLower(uriOrFlag)); f != nil {
			uriOrFlag = f.Value.String()
		}
	}

	// Create a copy of DefaultParserConfig and add default values from flags
	parserConfig := DefaultParserConfig
	parserConfig.DecodeHook = mapstructure.ComposeDecodeHookFunc(
		flagOverwriteHook(&flagValues),
		DefaultParserConfig.DecodeHook,
	)

	return &ConfOpt[T]{
		uri:        uriOrFlag,
		ParserConf: parserConfig,
	}
}

// LoadFlags parses struct fields and adds command-line flags for fields with 'usage' tag,
// parses command-line flags, maps flag values to the struct, and modifies the provided struct pointer
// This is a public entry function for flag parsing functionality
func LoadFlags[T any](result *T) error {
	return LoadFlagsCtx(context.Background(), result)
}

// LoadFlagsCtx parses struct fields and adds command-line flags for fields with 'usage' tag with context support,
// parses command-line flags, maps flag values to the struct, and modifies the provided struct pointer
// The context can be used for cancellation or timeout during flag parsing
func LoadFlagsCtx[T any](ctx context.Context, config *T) error {
	// Check if context is cancelled before starting
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Parse struct fields and add flags for fields with usage tag
	parseStructFlagsCtx[T](ctx)

	// Parse flags if not already parsed
	if !flag.Parsed() {
		flag.Parse()
	}

	// Check if context is cancelled after parsing
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Create a map to store flag values
	flagValues := make(map[string]any)

	// Visit all set flags and collect their values
	flag.Visit(func(f *flag.Flag) {
		flagValues[f.Name] = f.Value.String()
	})

	// Use mapstructure to decode flag values to struct
	decoderConfig := &mapstructure.DecoderConfig{
		Result:           config,
		WeaklyTypedInput: true,
		TagName:          "default",
	}

	mapDecoder, err := mapstructure.NewDecoder(decoderConfig)
	if err != nil {
		// If decoder creation fails, return error
		return fmt.Errorf("failed to create mapstructure decoder: %w", err)
	}

	// Decode flag values to struct
	if err := mapDecoder.Decode(flagValues); err != nil {
		// If decoding fails, return error
		return fmt.Errorf("failed to decode flag values to struct: %w", err)
	}

	return nil
}

// LoadWithFlags loads configuration from URI or flag and unmarshals it to the provided object
// If uriOrFlag is a valid URI, it will be used as the configuration source
// If uriOrFlag is not a URI, it will be treated as a flag name for configuration
// It also parses the struct type T and adds flags for fields with 'usage' tag
func LoadWithFlags[T any](obj *T, uriOrFlag string) error {
	return LoadWithFlagsCtx(context.Background(), obj, uriOrFlag)
}

// LoadWithFlagsCtx loads configuration from URI or flag with context and unmarshals it to the provided object
// If uriOrFlag is a valid URI, it will be used as the configuration source
// If uriOrFlag is not a URI, it will be treated as a flag name for configuration
// It also parses the struct type T and adds flags for fields with 'usage' tag
// The context can be used for cancellation or timeout during loading
func LoadWithFlagsCtx[T any](ctx context.Context, obj *T, uriOrFlag string) error {
	// Check if context is cancelled before starting
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Use NewWithFlagsCtx to create configuration option with flag support
	confOpt := NewWithFlagsCtx[T](ctx, uriOrFlag)

	// Load and decode configuration
	if err := confOpt.loadAndDecode(ctx); err != nil {
		return fmt.Errorf("failed to load and decode configuration: %w", err)
	}

	// Decode to the provided struct
	if err := confOpt.decodeToStruct(obj); err != nil {
		return fmt.Errorf("failed to decode to struct: %w", err)
	}

	return nil
}

// flagOverwriteHook creates a decode hook that overwrites config file values with flag values
func flagOverwriteHook[T any](flagValues *T) mapstructure.DecodeHookFuncType {
	return func(f reflect.Type, t reflect.Type, data any) (any, error) {
		// Always try to get flag value to overwrite config file value
		if flagVal := getFlagDefaultValue(flagValues, t); flagVal != nil {
			return flagVal, nil
		}

		// If no flag value found, return original data from config file
		return data, nil
	}
}

// getFlagDefaultValue extracts default value from flagValues struct for a given type
func getFlagDefaultValue[T any](flagValues *T, targetType reflect.Type) any {
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

// isValidURI checks if the given string is a valid URI
func isValidURI(str string) bool {
	if str == "" {
		return false
	}

	// Try to parse as URL
	u, err := url.Parse(str)
	if err != nil {
		return false
	}

	// Check if it has a scheme (http, https, file, etc.)
	return u.Scheme != ""
}

// parseStructFlagsCtx parses struct fields and adds flags for fields with usage tag with context support
// The context can be used for cancellation or timeout during flag parsing
func parseStructFlagsCtx[T any](ctx context.Context) {
	// Check if context is cancelled before starting
	select {
	case <-ctx.Done():
		return
	default:
	}

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
		// Check if context is cancelled during iteration
		select {
		case <-ctx.Done():
			return
		default:
		}

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
