package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sower-proxy/feconf"
	_ "github.com/sower-proxy/feconf/decoder/yaml"
	_ "github.com/sower-proxy/feconf/reader/http"
)

func main() {
	fmt.Println("=== HTTP + YAML Configuration Example ===")

	// Start config server in background
	fmt.Println("Starting config server...")
	go runServer()
	time.Sleep(2 * time.Second) // Wait for server to start

	// Run basic example
	fmt.Println("Running basic config example...")
	runBasicExample()

	// Run auth example
	fmt.Println("Running auth subscription example... (takes about 20 seconds)")
	runAuthExample()

	fmt.Println("All examples completed")
}

// Basic configuration loading example
func runBasicExample() {
	configURI := "http://localhost:8080/config.yaml"
	loader := feconf.New[Config]("", configURI)
	defer loader.Close()

	config, err := loader.Parse()
	if err != nil {
		fmt.Printf("Basic example failed: %v\n", err)
		return
	}

	fmt.Println("Basic config loaded successfully!")
	fmt.Printf("  Key config info:\n")
	fmt.Printf("    App: %s v%s (%s)\n",
		config.App.Name, config.App.Version, config.App.Environment)
	fmt.Printf("    Server: %s:%d (debug: %t)\n",
		config.Server.Host, config.Server.Port, config.Server.Debug)
	fmt.Printf("    Database: %s\n", config.Database.Primary.URL)
	fmt.Printf("    Replicas: %d\n", len(config.Database.Replicas))
	fmt.Printf("    Features: %v\n", config.Features)
	fmt.Println()
}

// Auth configuration subscription example
func runAuthExample() {
	configURI := "http://user:pass@localhost:8080/config-auth.yaml"
	loader := feconf.New[Config]("", configURI)
	defer loader.Close()

	fmt.Println("Starting auth config subscription...")

	// Subscribe to config changes
	eventChan, err := loader.Subscribe()
	if err != nil {
		fmt.Printf("Auth subscription failed: %v\n", err)
		fmt.Println("Hint: Make sure server supports auth (user:pass)")
		return
	}

	fmt.Println("Auth subscription successful!")
	fmt.Println("Listening for config changes... (expecting 3-4 events)")
	fmt.Println("  - Initial config event")
	fmt.Println("  - 3 update events (every 4 seconds)")
	fmt.Println()

	// Create timeout context - listen for 20 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	eventCount := 0
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("Auth subscription completed (received %d config events)\n", eventCount)
			return

		case event, ok := <-eventChan:
			if !ok {
				fmt.Println("Auth event channel closed")
				return
			}

			eventCount++
			if event.IsValid() {
				fmt.Printf("[Event %d] Auth config update successful\n", eventCount)
				fmt.Printf("  Time: %s\n", event.Timestamp.Format("15:04:05"))
				fmt.Printf("  Source: %s\n", event.SourceURI)
				fmt.Printf("  Key config changes:\n")
				fmt.Printf("    App: %s v%s (%s)\n",
					event.Config.App.Name, event.Config.App.Version, event.Config.App.Environment)
				fmt.Printf("    Server: %s:%d (debug: %t)\n",
					event.Config.Server.Host, event.Config.Server.Port, event.Config.Server.Debug)
				fmt.Printf("    Database: %s\n", event.Config.Database.Primary.URL)
				fmt.Printf("    Replicas: %d\n", len(event.Config.Database.Replicas))
				fmt.Printf("    Features: %v\n", event.Config.Features)
				fmt.Println("  " + strings.Repeat("-", 50))
			} else {
				fmt.Printf("[Event %d] Auth config update failed: %v\n", eventCount, event.Error)
			}
		}
	}
}
