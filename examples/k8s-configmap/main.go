package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/sower-proxy/conf"
	_ "github.com/sower-proxy/conf/decoder/json"
	_ "github.com/sower-proxy/conf/decoder/yaml"
	_ "github.com/sower-proxy/conf/reader/k8s"
)

// Config represents application configuration
type Config struct {
	Server ServerConfig `mapstructure:"server"`
	DB     DBConfig     `mapstructure:"database"`
}

// ServerConfig represents server configuration
type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

// DBConfig represents database configuration
type DBConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

func main() {
	// Example 1: Read from Kubernetes ConfigMap
	// URI format: k8s://{resourceType}/{namespace}/{name}[/{key}]
	// For example: k8s://configmap/default/app-config/config.yaml
	uri := "k8s://configmap/default/app-config/config.yaml"

	// Create a new configuration instance
	c := conf.New[Config](uri)
	defer c.Close()

	// Load configuration
	config, err := c.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Printf("Server config: %+v\n", config.Server)
	fmt.Printf("Database config: %+v\n", config.DB)

	// Example 2: Subscribe to configuration changes
	fmt.Println("\nSubscribing to configuration changes...")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Subscribe to configuration changes
	watchCtx, watchCancel := context.WithTimeout(ctx, 30*time.Second)
	defer watchCancel()

	events, err := c.SubscribeCtx(watchCtx)
	if err != nil {
		log.Fatalf("Failed to subscribe to config changes: %v", err)
	}

	// Listen for configuration updates
	for event := range events {
		if !event.IsValid() {
			log.Printf("Config update error: %v", event.Error)
			continue
		}

		fmt.Println("Configuration updated!")
		fmt.Printf("New server config: %+v\n", event.Config.Server)
		fmt.Printf("New database config: %+v\n", event.Config.DB)
	}

	fmt.Println("Subscription ended")
}