package main

import (
	"fmt"
	"log"
	"os"

	"github.com/sower-proxy/conf"
	_ "github.com/sower-proxy/conf/decoder/json"
	_ "github.com/sower-proxy/conf/reader/file"
	_ "github.com/sower-proxy/conf/reader/redis"
)

// Config 结构体展示了如何使用标签来定义命令行标志
// NewWithFlags 的参数是字段名，它会自动查找对应的标志
type Config struct {
	// ConfigURI 字段会自动生成 -configuri 标志
	// NewWithFlags("ConfigURI") 会自动查找 -configuri 标志
	// 如果标志未设置，则使用 "ConfigURI" 作为默认值
	ConfigURI string `usage:"Configuration file URI (file://, redis://, http://)" default:"file://./config.json"`

	// 应用程序基本配置
	AppName  string `usage:"Application name" default:"MyApp"`
	Version  string `usage:"Application version" default:"1.0.0"`
	LogLevel string `usage:"Log level (debug, info, warn, error)" default:"info"`

	// 服务器配置
	Server   ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`

	// 功能开关
	Features FeaturesConfig `json:"features"`
}

type ServerConfig struct {
	Host  string `usage:"Server host address" default:"localhost"`
	Port  int    `usage:"Server port number" default:"8080"`
	Debug bool   `usage:"Enable debug mode" default:"true"`
}

type DatabaseConfig struct {
	URL            string `usage:"Database connection URL" default:"postgresql://localhost:5432/myapp"`
	MaxConnections int    `usage:"Maximum database connections" default:"10"`
	Timeout        int    `usage:"Database timeout in seconds" default:"30"`
}

type FeaturesConfig struct {
	EnableCache    bool     `usage:"Enable caching feature" default:"true"`
	EnableMetrics  bool     `usage:"Enable metrics collection" default:"false"`
	AllowedOrigins []string `usage:"Allowed CORS origins (comma-separated)" default:"*"`
}

func main() {
	fmt.Println("🚀 Advanced Configuration Flags Example")
	fmt.Println("=====================================")

	// 首先加载标志以显示帮助信息
	flagValues, err := conf.LoadFlags[Config]()
	if err != nil {
		log.Printf("❌ Failed to parse flags: %v", err)
		os.Exit(1)
	}

	// 显示使用示例
	printUsageExamples()

	// 使用字段名作为参数创建配置加载器
	// NewWithFlags 会自动查找对应的标志（-configuri），如果没有设置则使用字段名作为默认值
	loader := conf.NewWithFlags[Config]("ConfigURI")
	defer loader.Close()

	// 加载配置
	config, err := loader.Load()
	if err != nil {
		log.Printf("❌ Failed to load configuration: %v", err)
		log.Println("💡 Make sure the configuration source is accessible and contains valid JSON")
		os.Exit(1)
	}

	fmt.Println("✅ Configuration loaded successfully!")
	fmt.Println("📋 Flag values used as defaults:")
	printFlagValues(*flagValues)
	fmt.Println()

	fmt.Println("📄 Final configuration:")
	printConfig(*config)
}

func printUsageExamples() {
	fmt.Println("💡 Usage examples:")
	fmt.Println("  # Basic usage with default config")
	fmt.Println("  go run main.go")
	fmt.Println()
	fmt.Println("  # Specify custom config file using -configuri flag")
	fmt.Println("  go run main.go -configuri file://./prod-config.json")
	fmt.Println()
	fmt.Println("  # Override individual settings via flags")
	fmt.Println("  go run main.go -appname \"MyApp\" -loglevel debug")
	fmt.Println()
	fmt.Println("  # Use Redis configuration")
	fmt.Println("  go run main.go -configuri redis://localhost:6379/app-config")
	fmt.Println()
	fmt.Println("  # Show help")
	fmt.Println("  go run main.go -help")
	fmt.Println()
	fmt.Println("  # NewWithFlags usage: parameter is field name 'ConfigURI'")
	fmt.Println("  # It automatically looks for -configuri flag, uses field name as fallback")
}

func printFlagValues(config Config) {
	fmt.Printf("   🏷️  App Name: %s\n", config.AppName)
	fmt.Printf("   📋 Version: %s\n", config.Version)
	fmt.Printf("   📊 Log Level: %s\n", config.LogLevel)
	fmt.Printf("   🔗 Config URI: %s\n", config.ConfigURI)
}

func printConfig(config Config) {
	fmt.Printf("📱 Application Info:\n")
	fmt.Printf("   🏷️  Name: %s\n", config.AppName)
	fmt.Printf("   📋 Version: %s\n", config.Version)
	fmt.Printf("   📊 Log Level: %s\n", config.LogLevel)
	fmt.Printf("   🔗 Config URI: %s\n", config.ConfigURI)
	fmt.Println()

	fmt.Printf("🖥️  Server Configuration:\n")
	fmt.Printf("   📍 Host: %s\n", config.Server.Host)
	fmt.Printf("   🔌 Port: %d\n", config.Server.Port)
	fmt.Printf("   🐛 Debug: %t\n", config.Server.Debug)
	fmt.Println()

	fmt.Printf("🗄️  Database Configuration:\n")
	fmt.Printf("   🔗 URL: %s\n", config.Database.URL)
	fmt.Printf("   📊 Max Connections: %d\n", config.Database.MaxConnections)
	fmt.Printf("   ⏱️  Timeout: %ds\n", config.Database.Timeout)
	fmt.Println()

	fmt.Printf("🚀 Features Configuration:\n")
	fmt.Printf("   💾 Cache Enabled: %t\n", config.Features.EnableCache)
	fmt.Printf("   📈 Metrics Enabled: %t\n", config.Features.EnableMetrics)
	fmt.Printf("   🌐 Allowed Origins: %v\n", config.Features.AllowedOrigins)
	fmt.Println()
}
