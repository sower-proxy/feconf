package main

import (
	"fmt"
	"log"
	"time"

	"github.com/sower-proxy/conf"
	_ "github.com/sower-proxy/conf/decoder/json"
	_ "github.com/sower-proxy/conf/reader/redis"
)

type DatabaseConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Database string `json:"database"`
	Username string `json:"username"`
}

func main() {
	fmt.Println("🔧 Redis Hash Field Configuration Example")
	fmt.Println("=========================================")

	// Example URI for Redis hash field
	// Format: redis://host:port/hash-key?content-type=format#field-name
	uri := "redis://localhost:6379/app-config?content-type=application/json#database"

	// Create configuration loader
	loader := conf.New[DatabaseConfig](uri)
	defer loader.Close()

	fmt.Printf("📍 Reading from: %s\n", uri)
	fmt.Println("   - Hash key: 'app-config'")
	fmt.Println("   - Hash field: 'database'")
	fmt.Println()

	// Load initial configuration
	fmt.Println("📥 Loading initial configuration...")
	config, err := loader.Load()
	if err != nil {
		log.Fatalf("❌ Failed to load config: %v", err)
	}

	fmt.Println("✅ Configuration loaded successfully!")
	printConfig(*config)

	// Subscribe to configuration changes
	fmt.Println("🔄 Subscribing to configuration changes...")
	fmt.Println("💡 To test changes, run:")
	fmt.Println("   redis-cli HSET app-config database '{\"host\":\"production.db\",\"port\":5432,\"database\":\"prod_db\",\"username\":\"prod_user\"}'")
	fmt.Println()

	eventChan, err := loader.Subscribe()
	if err != nil {
		log.Fatalf("❌ Failed to subscribe to changes: %v", err)
	}

	// Listen for configuration updates
	timeout := time.After(30 * time.Second)
	updateCount := 0

	for {
		select {
		case event := <-eventChan:
			if event.IsValid() {
				updateCount++
				fmt.Printf("🔥 Configuration Update #%d Detected!\n", updateCount)
				printConfig(*event.Config)
			} else if event.Error != nil {
				fmt.Printf("❌ Configuration error: %v\n", event.Error)
			}

		case <-timeout:
			fmt.Println("⏰ Example completed (timeout reached)")
			return
		}
	}
}

func printConfig(config DatabaseConfig) {
	fmt.Printf("🗄️  Database Configuration:\n")
	fmt.Printf("   📍 Host: %s\n", config.Host)
	fmt.Printf("   🔌 Port: %d\n", config.Port)
	fmt.Printf("   💾 Database: %s\n", config.Database)
	fmt.Printf("   👤 Username: %s\n", config.Username)
	fmt.Println()
}
