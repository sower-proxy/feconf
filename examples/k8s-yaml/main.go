package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/sower-proxy/conf"
	_ "github.com/sower-proxy/conf/decoder/yaml"
	_ "github.com/sower-proxy/conf/reader/file"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
}

type ServerConfig struct {
	Host  string `yaml:"host"`
	Port  int    `yaml:"port"`
	Debug bool   `yaml:"debug"`
}

type DatabaseConfig struct {
	URL            string `yaml:"url"`
	MaxConnections int    `yaml:"max_connections"`
}

type RedisConfig struct {
	Address  string `yaml:"address"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

func main() {
	// Load configuration from file (use absolute path with extension)
	loader := conf.New[Config]("file://./config.yaml")
	config, err := loader.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Printf("Server: %s:%d (debug: %t)\n",
		config.Server.Host, config.Server.Port, config.Server.Debug)
	fmt.Printf("Database: %s (max_conn: %d)\n",
		config.Database.URL, config.Database.MaxConnections)
	fmt.Printf("Redis: %s (db: %d)\n",
		config.Redis.Address, config.Redis.DB)

	fmt.Println()
	fmt.Println("=== Starting subscription to watch for config changes ===")

	// Subscribe to configuration changes
	eventChan, err := loader.Subscribe()
	if err != nil {
		log.Fatalf("Failed to subscribe to config changes: %v", err)
	}
	defer loader.Close()

	// Create a context with timeout for demonstration
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	go func() {
		for {
			select {
			case <-ctx.Done():
				fmt.Println("Subscription timeout, exiting...")
				return
			case event, ok := <-eventChan:
				if !ok {
					fmt.Println("Event channel closed")
					return
				}

				if event.IsValid() {
					fmt.Printf("[%s] Config updated from: %s\n",
						event.Timestamp.Format("2006-01-02 15:04:05"), event.SourceURI)
					fmt.Printf("  Server: %s:%d (debug: %t)\n",
						event.Config.Server.Host, event.Config.Server.Port, event.Config.Server.Debug)
					fmt.Printf("  Database: %s (max_conn: %d)\n",
						event.Config.Database.URL, event.Config.Database.MaxConnections)
					fmt.Printf("  Redis: %s (db: %d)\n",
						event.Config.Redis.Address, event.Config.Redis.DB)
				} else {
					fmt.Printf("[%s] Error: %v\n",
						event.Timestamp.Format("2006-01-02 15:04:05"), event.Error)
				}
			}
		}
	}()

	fmt.Println("Watching for configuration changes... (modify config.yaml to see updates)")
	fmt.Println("Press Ctrl+C to exit or wait for timeout")

	// Wait for context timeout or interruption
	<-ctx.Done()

}