package main

import (
	"fmt"
	"log"
	"time"

	"github.com/sower-proxy/conf"
	_ "github.com/sower-proxy/conf/decoder/ini"  // Import to register INI decoder
	_ "github.com/sower-proxy/conf/reader/redis" // Import to register Redis reader
)

// Helper function to format timestamp string
func formatTimestamp(ts string) string {
	if t, err := time.Parse(time.RFC3339, ts); err == nil {
		return t.Format("2006-01-02 15:04:05")
	}
	return ts // Return original if parsing fails
}

type Config struct {
	Timestamp string `json:"timestamp"` // Change to string for easier INI parsing
	Database  struct {
		Host    string `json:"host"`
		Port    int    `json:"port"`
		Name    string `json:"db_name"` // Different tag name
		Timeout int    `json:"timeout"`
	} `json:"database"`

	Server struct {
		Host  string `json:"bind_host"` // Different tag name
		Port  int    `json:"port"`
		Debug bool   `json:"debug_mode"` // Different tag name
	} `json:"server"`

	Logging struct {
		Level string `json:"log_level"` // Different tag name
		File  string `json:"file"`
	} `json:"logging"`
}

func main() {
	fmt.Println("Starting Redis INI configuration example...")

	// Load config from Redis with INI format
	c := conf.New[Config]("redis://localhost:6379/config?db=0&content-type=text/ini")

	// Load initial config
	cfg, err := c.Load()
	if err != nil {
		log.Printf("Failed to load config: %v", err)
		log.Println("Make sure Redis is running on localhost:6379 and key 'config' contains valid INI data")
		return
	}

	fmt.Printf("Initial Configuration:\n")
	fmt.Printf("  Timestamp: %s\n", formatTimestamp(cfg.Timestamp))
	fmt.Printf("  Database: %s:%d (name: %s)\n", cfg.Database.Host, cfg.Database.Port, cfg.Database.Name)
	fmt.Printf("  Server: %s:%d (debug: %t)\n", cfg.Server.Host, cfg.Server.Port, cfg.Server.Debug)
	fmt.Printf("  Logging: level=%s, file=%s\n", cfg.Logging.Level, cfg.Logging.File)
	fmt.Println()

	fmt.Println("🔄 Starting configuration monitoring (polling mode)")
	fmt.Println("📝 Checking for updates every 2 seconds...")
	fmt.Println("🧪 To test: run './setup.sh' in another terminal")
	fmt.Println()

	// Use polling approach (reliable)
	lastTimestamp := cfg.Timestamp
	eventCount := 0
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	timeout := time.After(30 * time.Second)

	for {
		select {
		case <-timeout:
			fmt.Println("✅ Example finished.")
			return

		case <-ticker.C:
			newCfg, err := c.Load()
			if err != nil {
				log.Printf("❌ Failed to reload config: %v", err)
				continue
			}

			// Check if config has changed (by timestamp)
			if newCfg.Timestamp != lastTimestamp {
				eventCount++
				fmt.Printf("🔥 === Config Update #%d Detected ===\n", eventCount)
				fmt.Printf("⏰ Old timestamp: %s\n", formatTimestamp(lastTimestamp))
				fmt.Printf("🆕 New timestamp: %s\n", formatTimestamp(newCfg.Timestamp))
				fmt.Printf("🏢 Server: %s:%d (debug: %t)\n", newCfg.Server.Host, newCfg.Server.Port, newCfg.Server.Debug)
				fmt.Printf("🗄️  Database: %s:%d (%s)\n", newCfg.Database.Host, newCfg.Database.Port, newCfg.Database.Name)
				fmt.Printf("📋 Logging: %s -> %s\n", newCfg.Logging.Level, newCfg.Logging.File)
				fmt.Println("✅ Configuration update applied successfully!")
				fmt.Println()

				lastTimestamp = newCfg.Timestamp
			}
		}
	}
}
