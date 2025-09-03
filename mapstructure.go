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

// StringToBoolHook 字符串到布尔值的钩子
func StringToBoolHook() mapstructure.DecodeHookFuncType {
	return func(f reflect.Type, t reflect.Type, data any) (any, error) {
		if f.Kind() != reflect.String || t.Kind() != reflect.Bool {
			return data, nil
		}

		switch strings.ToLower(strings.TrimSpace(data.(string))) {
		case "true", "yes", "1", "on", "enable", "enabled":
			return true, nil
		case "false", "no", "0", "off", "disable", "disabled", "":
			return false, nil
		default:
			return false, fmt.Errorf("cannot parse '%s' as boolean", data)
		}
	}
}

// StringToSlogLevelHook 字符串到slog.Level的钩子
func StringToSlogLevelHook() mapstructure.DecodeHookFuncType {
	return func(f reflect.Type, t reflect.Type, data any) (any, error) {
		if f.Kind() != reflect.String || t != reflect.TypeOf(slog.LevelDebug) {
			return data, nil
		}

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
