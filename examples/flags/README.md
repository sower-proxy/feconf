# Configuration Flags Example

This example demonstrates how to use the configuration library with command-line flags to specify configuration sources.

## Features

- ✅ **Command-line Flag Support** - Use `-config` flag to specify configuration URI
- ✅ **Default Fallback** - Fallback to default URI if no flag provided
- ✅ **Multiple Sources** - Works with files, Redis, HTTP, WebSocket sources
- ✅ **Custom Flag Names** - Support for custom flag names via `NewWithFlagsNamed`
- ✅ **Empty Value Handling** - Uses default if flag is provided but empty

## Quick Start

1. **Run with default configuration:**
   ```bash
   go run main.go
   # Uses: file://./config.json
   ```

2. **Run with custom file:**
   ```bash
   go run main.go -config file://./prod-config.json
   # Uses: file://./prod-config.json
   ```

3. **Run with Redis configuration:**
   ```bash
   # First setup Redis with configuration
   redis-cli SET app-config '{"server":{"host":"redis-app","port":8080,"debug":false},"database":{"url":"redis://localhost:6379/0","max_connections":20}}'
   
   # Run with Redis source
   go run main.go -config "redis://localhost:6379/app-config?content-type=application/json"
   ```

4. **Run with HTTP configuration:**
   ```bash
   go run main.go -config "https://config-server.example.com/app-config.json"
   ```

## API Usage

### Basic Flag Support
```go
// Uses default flag name "config"
loader := conf.NewWithFlags[Config]("file://./default-config.json")
```

### Custom Flag Name
```go
// Uses custom flag name "app-config"
loader := conf.NewWithFlagsNamed[Config]("app-config", "file://./default-config.json")
```

## Command Line Examples

```bash
# Default configuration
./app

# File-based configuration
./app -config file://./config.json
./app -config file:///absolute/path/to/config.yaml

# Redis configuration
./app -config "redis://localhost:6379/app-config?content-type=application/json"
./app -config "redis://localhost:6379/settings?db=1&content-type=application/json#database"

# HTTP configuration  
./app -config "https://config-server/app.json?timeout=30s"
./app -config "http://localhost:8080/config.yaml"

# WebSocket configuration
./app -config "wss://realtime-config.example.com/config?ping_interval=30s"

# Custom flag name
./app -app-config file://./app.toml
```

## Configuration Structure

The example uses this configuration structure:

```json
{
  "server": {
    "host": "localhost",
    "port": 8080,
    "debug": true
  },
  "database": {
    "url": "postgresql://localhost:5432/myapp", 
    "max_connections": 10
  }
}
```

## Expected Output

```
🚀 Configuration Flags Example
==============================
💡 Usage examples:
  go run main.go                                    # Uses default: file://./config.json
  go run main.go -config file://./prod-config.json  # Uses specified file
  go run main.go -config redis://localhost:6379/app # Uses Redis configuration

✅ Configuration loaded successfully!
🖥️  Server Configuration:
   📍 Host: localhost
   🔌 Port: 8080
   🐛 Debug: true
🗄️  Database Configuration:
   🔗 URL: postgresql://localhost:5432/myapp
   📊 Max Connections: 10
```

## Benefits

1. **Flexibility**: Switch between different configuration sources without code changes
2. **Environment Support**: Use different configs for dev/staging/prod environments
3. **Deployment Friendly**: Easy to configure in containers, systemd services, etc.
4. **Backward Compatible**: Falls back to default if no flag provided
5. **Validation**: Built-in URI validation and error handling

## Integration Examples

### Docker
```dockerfile
ENTRYPOINT ["./app", "-config", "redis://redis:6379/app-config?content-type=application/json"]
```

### Systemd Service
```ini
[Unit]
Description=My Application
After=network.target

[Service]
ExecStart=/usr/local/bin/app -config file:///etc/myapp/config.json
Restart=always
User=myapp

[Install]
WantedBy=multi-user.target
```

### Kubernetes
```yaml
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: app
        image: myapp:latest
        command: ["./app"]
        args: ["-config", "redis://redis-service:6379/app-config?content-type=application/json"]
```

## Troubleshooting

1. **"flag provided but not defined" error:**
   - Make sure to call `NewWithFlags` before accessing any flags
   - Avoid calling `flag.Parse()` manually in your application

2. **"Failed to load configuration" error:**
   - Verify the configuration URI is accessible
   - Check if the configuration source contains valid data
   - Ensure required decoders are imported (e.g., `_ "github.com/sower-proxy/conf/decoder/json"`)

3. **Flag not recognized:**
   - Make sure you're using the correct flag name (`-config` by default)
   - Use `NewWithFlagsNamed` for custom flag names