package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/sower-proxy/feconf"
	_ "github.com/sower-proxy/feconf/decoder/yaml"
	_ "github.com/sower-proxy/feconf/reader/k8s"
)

type Config struct {
	App      AppConfig      `yaml:"app"`
	Database DatabaseConfig `yaml:"database"`
	Features FeaturesConfig `yaml:"features"`
}

type AppConfig struct {
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	Environment string `yaml:"environment"`
	Debug       bool   `yaml:"debug"`
}

type DatabaseConfig struct {
	Host              string `yaml:"host"`
	Port              int    `yaml:"port"`
	Name              string `yaml:"name"`
	MaxConnections    int    `yaml:"max_connections"`
	ConnectionTimeout int    `yaml:"connection_timeout"`
}

type FeaturesConfig struct {
	EnableCache    bool     `yaml:"enable_cache"`
	EnableMetrics  bool     `yaml:"enable_metrics"`
	AllowedOrigins []string `yaml:"allowed_origins"`
}

func main() {
	// Load configuration from Kubernetes ConfigMap using ~/.kube/config
	// Format: k8s://[resource-type]/[namespace]/[name]/[key]
	// Example: k8s://configmap/default/my-app-config/config.yaml
	loader := feconf.New[Config]("", "k8s://configmap/default/my-app-config/config.yaml")
	defer loader.Close()

	config, err := loader.Parse()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Printf("=== Application Configuration ===\n")
	fmt.Printf("Name: %s\n", config.App.Name)
	fmt.Printf("Version: %s\n", config.App.Version)
	fmt.Printf("Environment: %s\n", config.App.Environment)
	fmt.Printf("Debug: %t\n", config.App.Debug)

	fmt.Printf("\n=== Database Configuration ===\n")
	fmt.Printf("Host: %s\n", config.Database.Host)
	fmt.Printf("Port: %d\n", config.Database.Port)
	fmt.Printf("Name: %s\n", config.Database.Name)
	fmt.Printf("Max Connections: %d\n", config.Database.MaxConnections)
	fmt.Printf("Connection Timeout: %ds\n", config.Database.ConnectionTimeout)

	fmt.Printf("\n=== Features Configuration ===\n")
	fmt.Printf("Enable Cache: %t\n", config.Features.EnableCache)
	fmt.Printf("Enable Metrics: %t\n", config.Features.EnableMetrics)
	fmt.Printf("Allowed Origins: %v\n", config.Features.AllowedOrigins)

	fmt.Println("\n=== Starting subscription to watch for config changes ===")

	// Subscribe to configuration changes
	eventChan, err := loader.Subscribe()
	if err != nil {
		log.Fatalf("Failed to subscribe to config changes: %v", err)
	}

	// Create a context with timeout for demonstration
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	go func() {
		for {
			select {
			case <-ctx.Done():
				fmt.Println("\nSubscription timeout, exiting...")
				return
			case event, ok := <-eventChan:
				if !ok {
					fmt.Println("\nEvent channel closed")
					return
				}

				if event.IsValid() {
					fmt.Printf("\n[%s] Config updated from: %s\n",
						event.Timestamp.Format("2006-01-02 15:04:05"), event.SourceURI)
					fmt.Printf("App Version: %s (Debug: %t)\n",
						event.Config.App.Version, event.Config.App.Debug)
					fmt.Printf("Database: %s:%d\n",
						event.Config.Database.Host, event.Config.Database.Port)
				} else {
					fmt.Printf("\n[%s] Error: %v\n",
						event.Timestamp.Format("2006-01-02 15:04:05"), event.Error)
				}
			}
		}
	}()

	fmt.Println("Watching for configuration changes...")
	fmt.Println("(Update the ConfigMap in Kubernetes to see real-time updates)")
	fmt.Println("Press Ctrl+C to exit or wait for timeout")

	// Wait for context timeout or interruption
	<-ctx.Done()
}
