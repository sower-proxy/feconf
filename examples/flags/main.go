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
// LoadWithFlags 的参数是字段名，它会自动查找对应的标志
type Config struct {
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
	var flagValues Config
	err := conf.LoadFlags(&flagValues)
	if err != nil {
		log.Printf("❌ Failed to parse flags: %v", err)
		os.Exit(1)
	}

	// 显示使用示例
	printUsageExamples()

	// 方法1：使用 LoadWithFlags 直接加载配置（推荐）
	// 这是更简洁的方式，一次性完成标志解析和配置加载
	var config Config
	err = conf.LoadWithFlags(&config, "ConfigURI")
	if err != nil {
		log.Printf("❌ Failed to load configuration: %v", err)
		log.Println("💡 Make sure the configuration source is accessible and contains valid JSON")
		os.Exit(1)
	}

	// 方法2：使用 ConfOpt 的 Load 方法（灵活方式）
	// 适合需要更多控制或复用配置选项的场景
	loader := conf.NewWithFlags[Config]("ConfigURI")
	defer loader.Close()

	var config2 Config
	err = loader.Load(&config2)
	if err != nil {
		log.Printf("❌ Failed to load configuration using ConfOpt: %v", err)
		os.Exit(1)
	}

	fmt.Println("✅ Configuration loaded successfully!")
	fmt.Println("📋 Flag values used as defaults:")
	printFlagValues(flagValues)
	fmt.Println()

	fmt.Println("📄 Final configuration:")
	printConfig(config)
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
	fmt.Println("  # LoadWithFlags usage: parameter is field name 'ConfigURI'")
	fmt.Println("  # It automatically looks for -configuri flag, uses field name as fallback")
	fmt.Println("  # This function combines flag parsing and config loading in one call")
	fmt.Println("  # Command line flags override config file values")
	fmt.Println()
	fmt.Println("  # ConfOpt.Load usage: create loader first, then load config")
	fmt.Println("  # This provides more control and allows reusing the loader")
	fmt.Println("  # Also supports command line flags and config file values")
}

func printFlagValues(config Config) {
	fmt.Printf("   🏷️  App Name: %s\n", config.AppName)
	fmt.Printf("   📋 Version: %s\n", config.Version)
	fmt.Printf("   📊 Log Level: %s\n", config.LogLevel)
}

func printConfig(config Config) {
	fmt.Printf("📱 Application Info:\n")
	fmt.Printf("   🏷️  Name: %s\n", config.AppName)
	fmt.Printf("   📋 Version: %s\n", config.Version)
	fmt.Printf("   📊 Log Level: %s\n", config.LogLevel)
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
