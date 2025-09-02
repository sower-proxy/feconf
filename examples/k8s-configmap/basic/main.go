package main

import (
	"fmt"
	"log"

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
	// This reads the 'config.yaml' key from the 'app-config' ConfigMap in the 'default' namespace
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

	fmt.Println("Basic example completed.")
}