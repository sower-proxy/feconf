package feconf

import (
	"log/slog"
	"reflect"
	"testing"

	"github.com/go-viper/mapstructure/v2"
)

func TestStringToBoolHook(t *testing.T) {
	hook := HookFuncStringToBool()

	tests := []struct {
		name     string
		fromType reflect.Type
		toType   reflect.Type
		data     any
		want     any
		wantErr  bool
	}{
		// 字符串测试
		{
			name:     "string true",
			fromType: reflect.TypeOf(""),
			toType:   reflect.TypeOf(true),
			data:     "true",
			want:     true,
			wantErr:  false,
		},
		{
			name:     "string false",
			fromType: reflect.TypeOf(""),
			toType:   reflect.TypeOf(true),
			data:     "false",
			want:     false,
			wantErr:  false,
		},
		{
			name:     "string 1",
			fromType: reflect.TypeOf(""),
			toType:   reflect.TypeOf(true),
			data:     "1",
			want:     true,
			wantErr:  false,
		},
		{
			name:     "string 0",
			fromType: reflect.TypeOf(""),
			toType:   reflect.TypeOf(true),
			data:     "0",
			want:     false,
			wantErr:  false,
		},
		{
			name:     "string invalid",
			fromType: reflect.TypeOf(""),
			toType:   reflect.TypeOf(true),
			data:     "invalid",
			want:     false,
			wantErr:  true,
		},

		// 整数测试
		{
			name:     "int 1",
			fromType: reflect.TypeOf(0),
			toType:   reflect.TypeOf(true),
			data:     1,
			want:     true,
			wantErr:  false,
		},
		{
			name:     "int 0",
			fromType: reflect.TypeOf(0),
			toType:   reflect.TypeOf(true),
			data:     0,
			want:     false,
			wantErr:  false,
		},
		{
			name:     "int8 -1",
			fromType: reflect.TypeOf(int8(0)),
			toType:   reflect.TypeOf(true),
			data:     int8(-1),
			want:     true,
			wantErr:  false,
		},
		{
			name:     "int16 100",
			fromType: reflect.TypeOf(int16(0)),
			toType:   reflect.TypeOf(true),
			data:     int16(100),
			want:     true,
			wantErr:  false,
		},
		{
			name:     "uint 1",
			fromType: reflect.TypeOf(uint(0)),
			toType:   reflect.TypeOf(true),
			data:     uint(1),
			want:     true,
			wantErr:  false,
		},
		{
			name:     "uint 0",
			fromType: reflect.TypeOf(uint(0)),
			toType:   reflect.TypeOf(true),
			data:     uint(0),
			want:     false,
			wantErr:  false,
		},

		// 浮点数测试
		{
			name:     "float32 1.0",
			fromType: reflect.TypeOf(float32(0)),
			toType:   reflect.TypeOf(true),
			data:     float32(1.0),
			want:     true,
			wantErr:  false,
		},
		{
			name:     "float32 0.0",
			fromType: reflect.TypeOf(float32(0)),
			toType:   reflect.TypeOf(true),
			data:     float32(0.0),
			want:     false,
			wantErr:  false,
		},
		{
			name:     "float64 -1.5",
			fromType: reflect.TypeOf(float64(0)),
			toType:   reflect.TypeOf(true),
			data:     float64(-1.5),
			want:     true,
			wantErr:  false,
		},
		{
			name:     "float64 0.0",
			fromType: reflect.TypeOf(float64(0)),
			toType:   reflect.TypeOf(true),
			data:     float64(0.0),
			want:     false,
			wantErr:  false,
		},

		// 非布尔目标类型测试
		{
			name:     "non-bool target",
			fromType: reflect.TypeOf(""),
			toType:   reflect.TypeOf(0),
			data:     "true",
			want:     "true",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := hook(tt.fromType, tt.toType, tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("StringToBoolHook() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("StringToBoolHook() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStringToBoolHookIntegration(t *testing.T) {
	type Config struct {
		Enabled bool `json:"enabled"`
	}

	tests := []struct {
		name     string
		input    map[string]any
		expected Config
		wantErr  bool
	}{
		{
			name:  "string true",
			input: map[string]any{"enabled": "true"},
			expected: Config{
				Enabled: true,
			},
			wantErr: false,
		},
		{
			name:  "string false",
			input: map[string]any{"enabled": "false"},
			expected: Config{
				Enabled: false,
			},
			wantErr: false,
		},
		{
			name:  "int 1",
			input: map[string]any{"enabled": 1},
			expected: Config{
				Enabled: true,
			},
			wantErr: false,
		},
		{
			name:  "int 0",
			input: map[string]any{"enabled": 0},
			expected: Config{
				Enabled: false,
			},
			wantErr: false,
		},
		{
			name:  "float 1.0",
			input: map[string]any{"enabled": 1.0},
			expected: Config{
				Enabled: true,
			},
			wantErr: false,
		},
		{
			name:  "float 0.0",
			input: map[string]any{"enabled": 0.0},
			expected: Config{
				Enabled: false,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var config Config
			decoderConfig := DefaultParserConfig
			decoderConfig.Result = &config
			decoder, err := mapstructure.NewDecoder(&decoderConfig)
			if err != nil {
				t.Fatalf("Failed to create decoder: %v", err)
			}

			err = decoder.Decode(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if config != tt.expected {
				t.Errorf("Decode() = %v, want %v", config, tt.expected)
			}
		})
	}
}

func TestStringToSlogLevelHook(t *testing.T) {
	hook := HookFuncStringToSlogLevel()

	tests := []struct {
		name     string
		fromType reflect.Type
		toType   reflect.Type
		data     any
		want     any
		wantErr  bool
	}{
		// 字符串测试
		{
			name:     "string debug",
			fromType: reflect.TypeOf(""),
			toType:   reflect.TypeOf(slog.LevelDebug),
			data:     "debug",
			want:     slog.LevelDebug,
			wantErr:  false,
		},
		{
			name:     "string info",
			fromType: reflect.TypeOf(""),
			toType:   reflect.TypeOf(slog.LevelDebug),
			data:     "info",
			want:     slog.LevelInfo,
			wantErr:  false,
		},
		{
			name:     "string warn",
			fromType: reflect.TypeOf(""),
			toType:   reflect.TypeOf(slog.LevelDebug),
			data:     "warn",
			want:     slog.LevelWarn,
			wantErr:  false,
		},
		{
			name:     "string error",
			fromType: reflect.TypeOf(""),
			toType:   reflect.TypeOf(slog.LevelDebug),
			data:     "error",
			want:     slog.LevelError,
			wantErr:  false,
		},
		{
			name:     "string -4",
			fromType: reflect.TypeOf(""),
			toType:   reflect.TypeOf(slog.LevelDebug),
			data:     "-4",
			want:     slog.LevelDebug,
			wantErr:  false,
		},
		{
			name:     "string 0",
			fromType: reflect.TypeOf(""),
			toType:   reflect.TypeOf(slog.LevelDebug),
			data:     "0",
			want:     slog.LevelInfo,
			wantErr:  false,
		},
		{
			name:     "string 4",
			fromType: reflect.TypeOf(""),
			toType:   reflect.TypeOf(slog.LevelDebug),
			data:     "4",
			want:     slog.LevelWarn,
			wantErr:  false,
		},
		{
			name:     "string 8",
			fromType: reflect.TypeOf(""),
			toType:   reflect.TypeOf(slog.LevelDebug),
			data:     "8",
			want:     slog.LevelError,
			wantErr:  false,
		},
		{
			name:     "string invalid",
			fromType: reflect.TypeOf(""),
			toType:   reflect.TypeOf(slog.LevelDebug),
			data:     "invalid",
			want:     nil,
			wantErr:  true,
		},

		// 数字测试
		{
			name:     "int -4",
			fromType: reflect.TypeOf(0),
			toType:   reflect.TypeOf(slog.LevelDebug),
			data:     -4,
			want:     slog.LevelDebug,
			wantErr:  false,
		},
		{
			name:     "int 0",
			fromType: reflect.TypeOf(0),
			toType:   reflect.TypeOf(slog.LevelDebug),
			data:     0,
			want:     slog.LevelInfo,
			wantErr:  false,
		},
		{
			name:     "int 4",
			fromType: reflect.TypeOf(0),
			toType:   reflect.TypeOf(slog.LevelDebug),
			data:     4,
			want:     slog.LevelWarn,
			wantErr:  false,
		},
		{
			name:     "int 8",
			fromType: reflect.TypeOf(0),
			toType:   reflect.TypeOf(slog.LevelDebug),
			data:     8,
			want:     slog.LevelError,
			wantErr:  false,
		},
		{
			name:     "int8 -4",
			fromType: reflect.TypeOf(int8(0)),
			toType:   reflect.TypeOf(slog.LevelDebug),
			data:     int8(-4),
			want:     slog.LevelDebug,
			wantErr:  false,
		},
		{
			name:     "uint 0",
			fromType: reflect.TypeOf(uint(0)),
			toType:   reflect.TypeOf(slog.LevelDebug),
			data:     uint(0),
			want:     slog.LevelInfo,
			wantErr:  false,
		},
		{
			name:     "float32 4.0",
			fromType: reflect.TypeOf(float32(0)),
			toType:   reflect.TypeOf(slog.LevelDebug),
			data:     float32(4.0),
			want:     slog.LevelWarn,
			wantErr:  false,
		},
		{
			name:     "float64 8.0",
			fromType: reflect.TypeOf(float64(0)),
			toType:   reflect.TypeOf(slog.LevelDebug),
			data:     float64(8.0),
			want:     slog.LevelError,
			wantErr:  false,
		},
		{
			name:     "int invalid",
			fromType: reflect.TypeOf(0),
			toType:   reflect.TypeOf(slog.LevelDebug),
			data:     1,
			want:     nil,
			wantErr:  true,
		},

		// 非slog.Level目标类型测试
		{
			name:     "non-slog.Level target",
			fromType: reflect.TypeOf(""),
			toType:   reflect.TypeOf(0),
			data:     "debug",
			want:     "debug",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := hook(tt.fromType, tt.toType, tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("StringToSlogLevelHook() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("StringToSlogLevelHook() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStringToSlogLevelHookIntegration(t *testing.T) {
	type Config struct {
		Level slog.Level `json:"level"`
	}

	tests := []struct {
		name     string
		input    map[string]any
		expected Config
		wantErr  bool
	}{
		{
			name:  "string debug",
			input: map[string]any{"level": "debug"},
			expected: Config{
				Level: slog.LevelDebug,
			},
			wantErr: false,
		},
		{
			name:  "string info",
			input: map[string]any{"level": "info"},
			expected: Config{
				Level: slog.LevelInfo,
			},
			wantErr: false,
		},
		{
			name:  "int -4",
			input: map[string]any{"level": -4},
			expected: Config{
				Level: slog.LevelDebug,
			},
			wantErr: false,
		},
		{
			name:  "int 0",
			input: map[string]any{"level": 0},
			expected: Config{
				Level: slog.LevelInfo,
			},
			wantErr: false,
		},
		{
			name:  "float 4.0",
			input: map[string]any{"level": 4.0},
			expected: Config{
				Level: slog.LevelWarn,
			},
			wantErr: false,
		},
		{
			name:  "float 8.0",
			input: map[string]any{"level": 8.0},
			expected: Config{
				Level: slog.LevelError,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var config Config
			decoderConfig := DefaultParserConfig
			decoderConfig.Result = &config
			decoder, err := mapstructure.NewDecoder(&decoderConfig)
			if err != nil {
				t.Fatalf("Failed to create decoder: %v", err)
			}

			err = decoder.Decode(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if config != tt.expected {
				t.Errorf("Decode() = %v, want %v", config, tt.expected)
			}
		})
	}
}

func TestDefaultHook(t *testing.T) {
	hook := HookFuncDefault()

	type testStruct struct {
		Name string
		Age  int
	}

	tests := []struct {
		name     string
		fromType reflect.Type
		toType   reflect.Type
		data     any
		want     any
		wantErr  bool
	}{
		// 字符串测试
		{
			name:     "empty string to string",
			fromType: reflect.TypeOf(""),
			toType:   reflect.TypeOf(""),
			data:     "",
			want:     "",
			wantErr:  false,
		},
		{
			name:     "non-empty string to string",
			fromType: reflect.TypeOf(""),
			toType:   reflect.TypeOf(""),
			data:     "hello",
			want:     "hello",
			wantErr:  false,
		},

		// 整数测试
		{
			name:     "zero int to int",
			fromType: reflect.TypeOf(0),
			toType:   reflect.TypeOf(0),
			data:     0,
			want:     int64(0),
			wantErr:  false,
		},
		{
			name:     "non-zero int to int",
			fromType: reflect.TypeOf(0),
			toType:   reflect.TypeOf(0),
			data:     42,
			want:     42,
			wantErr:  false,
		},

		// 浮点数测试
		{
			name:     "zero float to float",
			fromType: reflect.TypeOf(0.0),
			toType:   reflect.TypeOf(0.0),
			data:     0.0,
			want:     float64(0),
			wantErr:  false,
		},
		{
			name:     "non-zero float to float",
			fromType: reflect.TypeOf(0.0),
			toType:   reflect.TypeOf(0.0),
			data:     3.14,
			want:     3.14,
			wantErr:  false,
		},

		// 布尔值测试
		{
			name:     "false bool to bool",
			fromType: reflect.TypeOf(false),
			toType:   reflect.TypeOf(false),
			data:     false,
			want:     false,
			wantErr:  false,
		},
		{
			name:     "true bool to bool",
			fromType: reflect.TypeOf(false),
			toType:   reflect.TypeOf(false),
			data:     true,
			want:     true,
			wantErr:  false,
		},

		// nil 测试
		{
			name:     "nil to string",
			fromType: reflect.TypeOf(nil),
			toType:   reflect.TypeOf(""),
			data:     nil,
			want:     "",
			wantErr:  false,
		},
		{
			name:     "nil to int",
			fromType: reflect.TypeOf(nil),
			toType:   reflect.TypeOf(0),
			data:     nil,
			want:     int64(0),
			wantErr:  false,
		},
		{
			name:     "nil to bool",
			fromType: reflect.TypeOf(nil),
			toType:   reflect.TypeOf(false),
			data:     nil,
			want:     false,
			wantErr:  false,
		},

		// 结构体测试
		{
			name:     "zero struct to struct",
			fromType: reflect.TypeOf(testStruct{}),
			toType:   reflect.TypeOf(testStruct{}),
			data:     testStruct{},
			want:     testStruct{},
			wantErr:  false,
		},
		{
			name:     "non-zero struct to struct",
			fromType: reflect.TypeOf(testStruct{}),
			toType:   reflect.TypeOf(testStruct{}),
			data:     testStruct{Name: "John", Age: 30},
			want:     testStruct{Name: "John", Age: 30},
			wantErr:  false,
		},

		// 指针测试
		{
			name:     "nil pointer to pointer",
			fromType: reflect.TypeOf((*testStruct)(nil)),
			toType:   reflect.TypeOf((*testStruct)(nil)),
			data:     (*testStruct)(nil),
			want:     nil,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := hook(tt.fromType, tt.toType, tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("DefaultHook() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// 对于结构体，使用 DeepEqual 比较
			if tt.fromType != nil && (tt.fromType.Kind() == reflect.Struct || (tt.toType != nil && tt.toType.Kind() == reflect.Struct)) {
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("DefaultHook() = %v, want %v", got, tt.want)
				}
			} else {
				if got != tt.want {
					t.Errorf("DefaultHook() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestDefaultHookIntegration(t *testing.T) {
	type Config struct {
		Name    string `json:"name"`
		Age     int    `json:"age"`
		Enabled bool   `json:"enabled"`
		Level   int    `json:"level"`
	}

	tests := []struct {
		name     string
		input    map[string]any
		expected Config
		wantErr  bool
	}{
		{
			name:  "empty values should get defaults",
			input: map[string]any{"name": "", "age": 0, "enabled": false, "level": 0},
			expected: Config{
				Name:    "",
				Age:     0,
				Enabled: false,
				Level:   0,
			},
			wantErr: false,
		},
		{
			name:  "nil values should get defaults",
			input: map[string]any{"name": nil, "age": nil, "enabled": nil, "level": nil},
			expected: Config{
				Name:    "",
				Age:     0,
				Enabled: false,
				Level:   0,
			},
			wantErr: false,
		},
		{
			name:  "mixed zero and non-zero values",
			input: map[string]any{"name": "", "age": 25, "enabled": true, "level": 0},
			expected: Config{
				Name:    "",
				Age:     25,
				Enabled: true,
				Level:   0,
			},
			wantErr: false,
		},
		{
			name:  "all non-zero values should remain unchanged",
			input: map[string]any{"name": "John", "age": 30, "enabled": true, "level": 5},
			expected: Config{
				Name:    "John",
				Age:     30,
				Enabled: true,
				Level:   5,
			},
			wantErr: false,
		},
		{
			name:  "partial fields provided",
			input: map[string]any{"name": "Alice", "age": 28},
			expected: Config{
				Name:    "Alice",
				Age:     28,
				Enabled: false,
				Level:   0,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var config Config
			decoderConfig := DefaultParserConfig
			decoderConfig.Result = &config
			decoder, err := mapstructure.NewDecoder(&decoderConfig)
			if err != nil {
				t.Fatalf("Failed to create decoder: %v", err)
			}

			err = decoder.Decode(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if config != tt.expected {
				t.Errorf("Decode() = %v, want %v", config, tt.expected)
			}
		})
	}
}

func TestDefaultHookWithStruct(t *testing.T) {
	type Address struct {
		Street string `json:"street"`
		City   string `json:"city"`
	}

	type Person struct {
		Name    string  `json:"name"`
		Age     int     `json:"age"`
		Address Address `json:"address"`
	}

	tests := []struct {
		name     string
		input    map[string]any
		expected Person
		wantErr  bool
	}{
		{
			name: "empty struct should get zero values",
			input: map[string]any{
				"name": "",
				"age":  0,
				"address": map[string]any{
					"street": "",
					"city":   "",
				},
			},
			expected: Person{
				Name:    "",
				Age:     0,
				Address: Address{Street: "", City: ""},
			},
			wantErr: false,
		},
		{
			name: "partial struct fields",
			input: map[string]any{
				"name": "Bob",
				"age":  0,
				"address": map[string]any{
					"street": "123 Main St",
					"city":   "",
				},
			},
			expected: Person{
				Name:    "Bob",
				Age:     0,
				Address: Address{Street: "123 Main St", City: ""},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var config Person
			decoderConfig := DefaultParserConfig
			decoderConfig.Result = &config
			decoder, err := mapstructure.NewDecoder(&decoderConfig)
			if err != nil {
				t.Fatalf("Failed to create decoder: %v", err)
			}

			err = decoder.Decode(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if config != tt.expected {
				t.Errorf("Decode() = %v, want %v", config, tt.expected)
			}
		})
	}
}
