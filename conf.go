package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/pelletier/go-toml/v2"
	"github.com/sower-proxy/deferlog/v2"
	"gopkg.in/yaml.v3"
)

var Version, Date string

func ReadConfig[T interface{ Validate() error }](file string, conf *T) (err error) {
	defer func() { deferlog.DebugError(err, "ReadConfig", "file", file) }()

	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	decodeM := map[string]any{}
	switch filepath.Ext(file) {
	case ".json":
		err = json.NewDecoder(f).Decode(&decodeM)
	case ".toml":
		err = toml.NewDecoder(f).Decode(&decodeM)
	case ".yaml", ".yml":
		err = yaml.NewDecoder(f).Decode(&decodeM)
	}
	if err != nil {
		return fmt.Errorf("decode config(%s): %w", filepath.Base(file), err)
	}

	decoder, _ := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			renderEnvHook(),
			renderBoolHook(),
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
		),
		TagName: "json",
		Result:  conf,
		MatchName: func(mapKey, fieldName string) bool {
			return strings.EqualFold(strings.ReplaceAll(mapKey, "_", ""), fieldName)
		},
	})
	if err := decoder.Decode(decodeM); err != nil {
		return fmt.Errorf("mapstructure config: %w", err)
	}

	return (*conf).Validate()
}
func renderBoolHook() mapstructure.DecodeHookFuncType {
	return func(f reflect.Type, t reflect.Type, data any) (any, error) {
		if f.Kind() != reflect.String || t.Kind() != reflect.Bool {
			return data, nil
		}

		switch strings.ToLower(strings.TrimSpace(data.(string))) {
		case "true", "yes", "1", "on", "enable", "enabled":
			return true, nil
		case "false", "no", "0", "off", "disable", "disabled":
			return false, nil
		default:
			return false, fmt.Errorf("cannot parse '%s' as boolean", data)
		}
	}
}

func renderEnvHook() mapstructure.DecodeHookFuncType {
	return func(f reflect.Type, t reflect.Type, data any) (any, error) {
		if f.Kind() != reflect.String {
			return data, nil
		}

		return renderEnv(data.(string)), nil
	}
}

var envRe = regexp.MustCompile(`\$\{([a-zA-Z0-9_]+)(?::-([^}]*))?\}`)

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

		// 检查是否被转义（前面有$符号）
		if prevByte(value, start) == '$' {
			// 保持转义的内容，但去掉一个$符号
			result += value[lastEnd:start-1] + value[start:end]
			lastEnd = end
			continue
		}

		// 添加变量前的内容
		result += value[lastEnd:start]

		// 处理环境变量
		envName := match[1]
		defaultValue := match[2] // 可能为空字符串
		envValue := os.Getenv(envName)

		// 如果环境变量不存在且有默认值，使用默认值
		if envValue == "" && defaultValue != "" {
			envValue = defaultValue
		}

		result += envValue
		lastEnd = end
	}

	// 添加最后剩余的内容
	result += value[lastEnd:]
	return result
}

func prevByte(value string, idx int) byte {
	if idx == 0 {
		return 0
	}
	return value[idx-1]
}
