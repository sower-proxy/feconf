package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/sower-proxy/feconf"
	_ "github.com/sower-proxy/feconf/decoder/xml"
	_ "github.com/sower-proxy/feconf/reader/file"
)

type Config struct {
	Server struct {
		Host string `xml:"host"`
		Port int    `xml:"port"`
	} `xml:"server"`
	Database struct {
		URL string `xml:"url"`
	} `xml:"database"`
}

func main() {
	fmt.Println("=== File + XML Configuration Example ===")

	// Get current directory for config file path
	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current directory: %v", err)
	}
	configPath := filepath.Join(currentDir, "config.xml")
	configURI := "file://" + configPath

	// Create the configuration loader
	// New signature: New[T](flag string, uris ...string)
	loader := feconf.New[Config]("", configURI)
	defer loader.Close()

	fmt.Println("Loading configuration...")
	config, err := loader.Parse()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	fmt.Printf("Configuration loaded: %s:%d, DB: %s\n",
		config.Server.Host, config.Server.Port, config.Database.URL)

	// Subscribe to configuration changes
	fmt.Println("Subscribing to configuration updates...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	eventChan, err := loader.Subscribe()
	if err != nil {
		log.Fatalf("Failed to subscribe: %v", err)
	}

	fmt.Println("Subscription successful, watching for changes...")
	fmt.Println("Try editing config.xml to see live updates!")
	fmt.Println()

	eventCount := 0
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("Watch ended (received %d events)\n", eventCount)
			return
		case event, ok := <-eventChan:
			if !ok {
				fmt.Println("Event channel closed")
				return
			}
			eventCount++
			if event.IsValid() {
				fmt.Printf("[Event %d] Config updated: %s:%d, DB: %s\n", eventCount,
					event.Config.Server.Host, event.Config.Server.Port, event.Config.Database.URL)
			} else {
				fmt.Printf("[Event %d] Config update failed: %v\n", eventCount, event.Error)
			}
		}
	}
}
