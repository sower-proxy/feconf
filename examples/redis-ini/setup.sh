#!/bin/bash

# Redis INI Example Setup Script

echo "Setting up Redis INI configuration example..."

# Check if Redis is running
if ! command -v redis-cli &> /dev/null; then
    echo "Redis CLI not found. Please install Redis server."
    echo "On Ubuntu/Debian: sudo apt-get install redis-server"
    echo "On macOS: brew install redis"
    exit 1
fi

# Test Redis connection
if ! redis-cli ping &> /dev/null; then
    echo "Redis server is not running. Please start Redis server:"
    echo "sudo systemctl start redis"
    echo "Or: redis-server"
    exit 1
fi

echo "Redis is running. Configuring keyspace notifications..."

# Enable keyspace notifications for key changes
# K = keyspace events, s = string commands, $ = generic commands  
redis-cli CONFIG SET notify-keyspace-events "K$"
echo "Keyspace notifications enabled (K$)"

echo "Loading INI config with current timestamp..."

# Generate current timestamp in ISO 8601 format
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

# Create config with current timestamp
redis-cli SET config "timestamp = $TIMESTAMP

[database]
host = localhost
port = 5432
db_name = myapp
timeout = 30

[server]
bind_host = 0.0.0.0
port = 8080
debug_mode = true

[logging]
log_level = info
file = /var/log/app.log"

echo "Configuration loaded into Redis key 'config' with timestamp: $TIMESTAMP"
echo "You can now run: go run main.go"

# Show current config
echo ""
echo "Current config in Redis:"
redis-cli GET config

echo ""
echo "Keyspace notification config:"
redis-cli CONFIG GET notify-keyspace-events