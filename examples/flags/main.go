package main

import (
	"fmt"
	"log"
	"os"

	"github.com/sower-proxy/feconf"
	_ "github.com/sower-proxy/feconf/decoder/json"
	_ "github.com/sower-proxy/feconf/reader/file"
)

// Config demonstrates how to use usage tags to define command-line flags
// Fields with usage tags will automatically be registered as flags
type Config struct {
	// Application basic configuration - these will become flags
	AppName  string `json:"app_name" usage:"Application name"`
	LogLevel string `json:"log_level" usage:"Log level (debug, info, warn, error)"`

	// Server configuration
	Server ServerConfig `json:"server"`

	// Database configuration
	Database DatabaseConfig `json:"database"`
}

type ServerConfig struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

type DatabaseConfig struct {
	URL            string `json:"url"`
	MaxConnections int    `json:"max_connections"`
}

func main() {
	fmt.Println("Configuration with Flags Example")
	fmt.Println("=================================")

	printUsageExamples()

	// Get config file path from current directory
	configPath, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current directory: %v", err)
	}
	configURI := "file://" + configPath + "/config.json"

	// Create a new configuration loader
	// New signature: New[T](flag string, uris ...string)
	// - flag: "config" means a -config flag will be added to override URI
	// - uris: default config file URIs
	// - Fields with usage tags will be registered as command-line flags
	// - Flag values override config file values
	loader := feconf.New[Config]("config", configURI)
	defer loader.Close()

	// Parse the configuration
	// This will:
	// 1. Parse command-line flags
	// 2. Load configuration from the file
	// 3. Merge flag values (flags take precedence)
	config, err := loader.Parse()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	fmt.Println("Configuration loaded successfully!")
	fmt.Println()
	printConfig(*config)
}

func printUsageExamples() {
	fmt.Println("Usage examples:")
	fmt.Println("  # Basic usage with config file")
	fmt.Println("  go run main.go")
	fmt.Println()
	fmt.Println("  # Override settings via flags (fields with usage tag)")
	fmt.Println("  go run main.go -app_name \"CustomApp\" -log_level debug")
	fmt.Println()
	fmt.Println("  # Use -config flag to specify config file")
	fmt.Println("  go run main.go -config file://./other-config.json")
	fmt.Println()
	fmt.Println("  # Show help")
	fmt.Println("  go run main.go -help")
	fmt.Println()
}

func printConfig(config Config) {
	fmt.Println("Application Info:")
	fmt.Printf("  Name: %s\n", config.AppName)
	fmt.Printf("  Log Level: %s\n", config.LogLevel)
	fmt.Println()

	fmt.Println("Server Configuration:")
	fmt.Printf("  Host: %s\n", config.Server.Host)
	fmt.Printf("  Port: %d\n", config.Server.Port)
	fmt.Println()

	fmt.Println("Database Configuration:")
	fmt.Printf("  URL: %s\n", config.Database.URL)
	fmt.Printf("  Max Connections: %d\n", config.Database.MaxConnections)
}
