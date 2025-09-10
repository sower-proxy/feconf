package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/sower-proxy/feconf"
	_ "github.com/sower-proxy/feconf/decoder/json"
	_ "github.com/sower-proxy/feconf/reader/file"
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
	// Load configuration from file (use absolute path with extension)
	loader := conf.New[Config]("file://./config.json")
	var config Config
	err := loader.Load(&config)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Printf("Server: %s:%d (debug: %t)\n",
		config.Server.Host, config.Server.Port, config.Server.Debug)
	fmt.Printf("Database: %s (max_conn: %d)\n",
		config.Database.URL, config.Database.MaxConnections)

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
				} else {
					fmt.Printf("[%s] Error: %v\n",
						event.Timestamp.Format("2006-01-02 15:04:05"), event.Error)
				}
			}
		}
	}()

	fmt.Println("Watching for configuration changes... (modify config.json to see updates)")
	fmt.Println("Press Ctrl+C to exit or wait for timeout")

	// Wait for context timeout or interruption
	<-ctx.Done()
}
