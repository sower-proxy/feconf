package main

import (
	"fmt"
	"log"

	"github.com/sower-proxy/conf"
	_ "github.com/sower-proxy/conf/decoder/json"
	_ "github.com/sower-proxy/conf/reader/file"
	_ "github.com/sower-proxy/conf/reader/redis"
)

type Config struct {
	Server   ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
}

type ServerConfig struct {
	Host  string `json:"host"`
	Port  int    `json:"port"`
	Debug bool   `json:"debug"`
}

type DatabaseConfig struct {
	URL            string `json:"url"`
	MaxConnections int    `json:"max_connections"`
}

func main() {
	fmt.Println("🚀 Configuration Flags Example")
	fmt.Println("==============================")

	// Use NewWithFlags to read config URI from command-line flags
	// Default to local file if no -config flag is provided
	loader := conf.NewWithFlags[Config]("file://./config.json")
	defer loader.Close()

	fmt.Println("💡 Usage examples:")
	fmt.Println("  go run main.go                                    # Uses default: file://./config.json")
	fmt.Println("  go run main.go -config file://./prod-config.json  # Uses specified file")
	fmt.Println("  go run main.go -config redis://localhost:6379/app # Uses Redis configuration")
	fmt.Println()

	// Load configuration
	config, err := loader.Load()
	if err != nil {
		log.Printf("❌ Failed to load configuration: %v", err)
		log.Println("💡 Make sure the configuration source is accessible and contains valid JSON")
		return
	}

	fmt.Println("✅ Configuration loaded successfully!")
	printConfig(*config)
}

func printConfig(config Config) {
	fmt.Printf("🖥️  Server Configuration:\n")
	fmt.Printf("   📍 Host: %s\n", config.Server.Host)
	fmt.Printf("   🔌 Port: %d\n", config.Server.Port)
	fmt.Printf("   🐛 Debug: %t\n", config.Server.Debug)
	fmt.Printf("🗄️  Database Configuration:\n")
	fmt.Printf("   🔗 URL: %s\n", config.Database.URL)
	fmt.Printf("   📊 Max Connections: %d\n", config.Database.MaxConnections)
	fmt.Println()
}