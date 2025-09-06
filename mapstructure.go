package conf

import (
	"fmt"
	"log/slog"
	"os"
	"reflect"
	"regexp"
	"strings"

	"github.com/go-viper/mapstructure/v2"
)

// DefaultParserConfig 默认解析器配置
var DefaultParserConfig = mapstructure.DecoderConfig{
	DecodeHook: mapstructure.ComposeDecodeHookFunc(
		DefaultHook(),
		EnvRenderHook(),
		StringToBoolHook(),
		StringToSlogLevelHook(),
		mapstructure.StringToTimeDurationHookFunc(),
		mapstructure.StringToSliceHookFunc(","),
		mapstructure.StringToBasicTypeHookFunc(),
	),
	TagName:          "json",
	WeaklyTypedInput: true,
	ErrorUnused:      false,
	ZeroFields:       false,
	MatchName: func(mapKey, fieldName string) bool {
		return strings.EqualFold(strings.ReplaceAll(mapKey, "_", ""), fieldName)
	},
}

// StringToBoolHook 字符串和数字到布尔值的钩子
func StringToBoolHook() mapstructure.DecodeHookFuncType {
	return func(f reflect.Type, t reflect.Type, data any) (any, error) {
		if t.Kind() != reflect.Bool {
			return data, nil
		}

		// 处理字符串类型
		if f.Kind() == reflect.String {
			switch strings.ToLower(strings.TrimSpace(data.(string))) {
			case "true", "yes", "1", "on", "enable", "enabled":
				return true, nil
			case "false", "no", "0", "off", "disable", "disabled", "":
				return false, nil
			default:
				return false, fmt.Errorf("cannot parse '%s' as boolean", data)
			}
		}

		// 处理数字类型
		if f.Kind() >= reflect.Int && f.Kind() <= reflect.Complex128 {
			// 使用反射来处理所有数字类型
			val := reflect.ValueOf(data)
			switch val.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				return val.Int() != 0, nil
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				return val.Uint() != 0, nil
			case reflect.Float32, reflect.Float64:
				return val.Float() != 0, nil
			case reflect.Complex64, reflect.Complex128:
				return val.Complex() != 0, nil
			}
		}

		return data, nil
	}
}

// StringToSlogLevelHook 字符串和数字到slog.Level的钩子
func StringToSlogLevelHook() mapstructure.DecodeHookFuncType {
	return func(f reflect.Type, t reflect.Type, data any) (any, error) {
		if t != reflect.TypeOf(slog.LevelDebug) {
			return data, nil
		}

		// 处理字符串类型
		if f.Kind() == reflect.String {
			level := strings.ToLower(strings.TrimSpace(data.(string)))
			switch level {
			case "debug", "dbg", "-4":
				return slog.LevelDebug, nil
			case "info", "information", "0":
				return slog.LevelInfo, nil
			case "warn", "warning", "4":
				return slog.LevelWarn, nil
			case "error", "err", "8":
				return slog.LevelError, nil
			default:
				return nil, fmt.Errorf("cannot parse '%s' as slog.Level", data)
			}
		}

		// 处理数字类型
		if f.Kind() >= reflect.Int && f.Kind() <= reflect.Complex128 {
			val := reflect.ValueOf(data)
			var levelValue int64

			switch val.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				levelValue = val.Int()
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				levelValue = int64(val.Uint())
			case reflect.Float32, reflect.Float64:
				levelValue = int64(val.Float())
			case reflect.Complex64, reflect.Complex128:
				levelValue = int64(real(val.Complex()))
			}

			// 根据数字值返回对应的 slog.Level
			switch levelValue {
			case -4:
				return slog.LevelDebug, nil
			case 0:
				return slog.LevelInfo, nil
			case 4:
				return slog.LevelWarn, nil
			case 8:
				return slog.LevelError, nil
			default:
				return nil, fmt.Errorf("cannot parse '%v' as slog.Level", data)
			}
		}

		return data, nil
	}
}

// EnvRenderHook 环境变量渲染钩子
func EnvRenderHook() mapstructure.DecodeHookFuncType {
	return func(f reflect.Type, t reflect.Type, data any) (any, error) {
		if f.Kind() != reflect.String {
			return data, nil
		}

		return renderEnv(data.(string)), nil
	}
}

var envRe = regexp.MustCompile(`\$\{([a-zA-Z0-9_]+)(?::-([^}]*))?\}`)

// renderEnv 渲染环境变量
func renderEnv(value string) string {
	matches := envRe.FindAllStringSubmatch(value, -1)
	idxPairs := envRe.FindAllStringIndex(value, -1)
	if len(matches) == 0 {
		return value
	}

	result := ""
	lastEnd := 0

	for i, match := range matches {
		idxPair := idxPairs[i]
		start := idxPair[0]
		end := idxPair[1]

		// 检查是否转义 ($${})
		if start > 0 && value[start-1] == '$' {
			result += value[lastEnd:start-1] + value[start:end]
			lastEnd = end
			continue
		}

		result += value[lastEnd:start]

		envName := match[1]
		defaultValue := match[2]
		envValue := os.Getenv(envName)

		if envValue == "" && defaultValue != "" {
			envValue = defaultValue
		}

		result += envValue
		lastEnd = end
	}

	result += value[lastEnd:]
	return result
}

// DefaultHook 默认值钩子，当其他钩子都无法处理时提供默认值
func DefaultHook() mapstructure.DecodeHookFuncType {
	return func(f reflect.Type, t reflect.Type, data any) (any, error) {
		// 如果数据为零值，根据目标类型返回默认值
		if isZeroValue(data) {
			return getDefaultValue(t), nil
		}

		// 如果数据不是零值，直接返回
		return data, nil
	}
}

// isZeroValue 检查值是否为零值
func isZeroValue(data any) bool {
	if data == nil {
		return true
	}

	val := reflect.ValueOf(data)
	switch val.Kind() {
	case reflect.String:
		return val.String() == ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return val.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return val.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return val.Float() == 0
	case reflect.Bool:
		return !val.Bool()
	case reflect.Ptr, reflect.Slice, reflect.Map, reflect.Interface:
		return val.IsNil()
	case reflect.Struct:
		// 对于结构体，检查所有字段是否为零值
		return reflect.DeepEqual(data, reflect.Zero(val.Type()).Interface())
	default:
		return false
	}
}

// getDefaultValue 根据类型返回默认值
func getDefaultValue(t reflect.Type) any {
	switch t.Kind() {
	case reflect.String:
		return ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return int64(0)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return uint64(0)
	case reflect.Float32, reflect.Float64:
		return float64(0)
	case reflect.Bool:
		return false
	case reflect.Ptr, reflect.Slice, reflect.Map, reflect.Interface:
		return nil
	case reflect.Struct:
		// 对于结构体，返回零值
		return reflect.Zero(t).Interface()
	default:
		return nil
	}
}
