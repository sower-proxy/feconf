# Redis INI Configuration Example

This example demonstrates how to use the configuration library with Redis as the data source and INI format.

## Features

- ✅ **Real-time configuration monitoring** using polling approach
- ✅ **Timestamp-based change detection** for reliable updates
- ✅ **Automatic configuration reload** when changes are detected
- ✅ **User-friendly output** with emojis and clear formatting
- ✅ **Robust error handling** with graceful degradation

## Prerequisites

1. Redis server running on localhost:6379
2. Go environment with required dependencies

## Quick Start

1. **Setup the configuration:**
   ```bash
   ./setup.sh
   ```

2. **Run the example:**
   ```bash
   go run main.go
   ```

3. **Test configuration updates:**
   ```bash
   # In another terminal while main.go is running
   ./setup.sh
   ```

## How It Works

The example uses a **polling-based approach** to monitor configuration changes:

1. **Loads initial configuration** from Redis key `config`
2. **Polls every 2 seconds** for configuration changes
3. **Detects changes** by comparing timestamps
4. **Displays updates** with detailed information about what changed
5. **Continues monitoring** until timeout (30 seconds)

## Configuration Structure

The INI configuration includes:

```ini
timestamp = 2025-08-29T15:48:21Z

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
file = /var/log/app.log
```

## Test Configuration Updates

### Method 1: Automated setup
```bash
./setup.sh  # Updates timestamp automatically
```

### Method 2: Multiple test updates
```bash
./test-updates.sh  # Triggers several updates with different values
```

### Method 3: Manual updates
```bash
redis-cli SET config "timestamp = $(date -u +"%Y-%m-%dT%H:%M:%SZ")

[database]
host = localhost
port = 5432
db_name = production_db
timeout = 60

[server]
bind_host = 0.0.0.0
port = 9000
debug_mode = false

[logging]
log_level = warn
file = /var/log/prod.log"
```

## Expected Output

```
Starting Redis INI configuration example...
Initial Configuration:
  Timestamp: 2025-08-29 15:47:55
  Database: localhost:5432 (name: myapp)
  Server: 0.0.0.0:8080 (debug: true)
  Logging: level=info, file=/var/log/app.log

🔄 Starting configuration monitoring (polling mode)
📝 Checking for updates every 2 seconds...
🧪 To test: run './setup.sh' in another terminal

🔥 === Config Update #1 Detected ===
⏰ Old timestamp: 2025-08-29 15:47:55
🆕 New timestamp: 2025-08-29 15:48:21
🏢 Server: 0.0.0.0:8080 (debug: true)
🗄️  Database: localhost:5432 (myapp)
📋 Logging: info -> /var/log/app.log
✅ Configuration update applied successfully!
```

## Why Polling Instead of Subscription?

This example uses **polling** rather than Redis keyspace notifications because:

- **Reliability**: Polling works consistently across different Redis configurations
- **Simplicity**: No need to configure keyspace notifications
- **Compatibility**: Works with Redis, Valkey, and other Redis-compatible servers
- **Predictability**: Updates are checked at regular intervals

For production use, you might want to:
- Increase polling interval to reduce Redis load
- Add exponential backoff on errors
- Implement proper logging
- Add configuration validation

## Troubleshooting

1. **"Failed to load config" error:**
   - Ensure Redis is running: `redis-cli ping`
   - Check if key exists: `redis-cli GET config`
   - Run setup script: `./setup.sh`

2. **No configuration updates detected:**
   - Verify timestamp is changing: `redis-cli GET config`
   - Check if setup script runs without errors
   - Ensure main.go is still running

3. **Connection refused error:**
   - Start Redis server: `sudo systemctl start redis`
   - Or run Redis in foreground: `redis-server`