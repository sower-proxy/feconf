package main

import (
	"fmt"
	"log"

	"github.com/sower-proxy/conf"
	_ "github.com/sower-proxy/conf/decoder/json"
	_ "github.com/sower-proxy/conf/decoder/yaml"
	_ "github.com/sower-proxy/conf/reader/k8s"
)

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

// Config represents application configuration
type Config struct {
	Server ServerConfig `mapstructure:"server"`
	DB     DBConfig     `mapstructure:"database"`
}

func main() {
	// Example 1: Read from a ConfigMap with multiple keys
	// This will read the first key's value if there's only one key,
	// or return an error if there are multiple keys and no specific key is specified
	fmt.Println("=== Reading from ConfigMap ===")
	uri1 := "k8s://configmap/default/multi-key-config"
	c1 := conf.New[Config](uri1)
	defer c1.Close()

	config1, err := c1.Load()
	if err != nil {
		log.Printf("Failed to load config from multi-key ConfigMap: %v", err)
	} else {
		fmt.Printf("Server config: %+v\n", config1.Server)
		fmt.Printf("Database config: %+v\n", config1.DB)
	}

	// Example 2: Read a specific key from a ConfigMap
	// This reads only the 'app.yaml' key from the ConfigMap
	fmt.Println("\n=== Reading specific key from ConfigMap ===")
	uri2 := "k8s://configmap/default/multi-key-config/app.yaml"
	c2 := conf.New[Config](uri2)
	defer c2.Close()

	config2, err := c2.Load()
	if err != nil {
		log.Printf("Failed to load config from specific key in ConfigMap: %v", err)
	} else {
		fmt.Printf("Server config: %+v\n", config2.Server)
		fmt.Printf("Database config: %+v\n", config2.DB)
	}

	// Example 3: Read from a Secret with multiple keys
	// This will read the first key's value if there's only one key,
	// or return an error if there are multiple keys and no specific key is specified
	fmt.Println("\n=== Reading from Secret ===")
	uri3 := "k8s://secret/default/multi-key-secret"
	c3 := conf.New[string](uri3)
	defer c3.Close()

	secret1, err := c3.Load()
	if err != nil {
		log.Printf("Failed to load secret from multi-key Secret: %v", err)
	} else {
		fmt.Printf("Secret value: %s\n", *secret1)
	}

	// Example 4: Read a specific key from a Secret
	// This reads only the 'password' key from the Secret
	fmt.Println("\n=== Reading specific key from Secret ===")
	uri4 := "k8s://secret/default/multi-key-secret/password"
	c4 := conf.New[string](uri4)
	defer c4.Close()

	secret2, err := c4.Load()
	if err != nil {
		log.Printf("Failed to load secret from specific key in Secret: %v", err)
	} else {
		fmt.Printf("Password: %s\n", *secret2)
	}

	// Example 5: Read another specific key from a Secret
	// This reads only the 'api-key' key from the Secret
	fmt.Println("\n=== Reading another specific key from Secret ===")
	uri5 := "k8s://secret/default/multi-key-secret/api-key"
	c5 := conf.New[string](uri5)
	defer c5.Close()

	secret3, err := c5.Load()
	if err != nil {
		log.Printf("Failed to load secret from specific key in Secret: %v", err)
	} else {
		fmt.Printf("API Key: %s\n", *secret3)
	}

	fmt.Println("\nAdvanced examples completed.")
}