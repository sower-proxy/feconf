# Advanced Configuration Flags Example

This example demonstrates the advanced flag functionality of the configuration library, including automatic flag generation from struct fields, default values, and type-safe configuration.

## 🚀 New Features

- ✅ **Automatic Flag Generation** - Struct fields with `usage` tag automatically generate command-line flags
- ✅ **Type-Safe Flags** - Support for string, bool, int, uint, float64, and string slice types
- ✅ **Default Values** - Set default values using `default` tag
- ✅ **Flag as Config Defaults** - Flag values are used as defaults when loading configuration
- ✅ **Help Text** - Automatic help generation from `usage` tags
- ✅ **Flexible Configuration** - Mix file-based config with command-line overrides

## 🏗️ Struct Tag Definition

```go
type Config struct {
    // Basic flags with usage and default values
    AppName  string `usage:"Application name" default:"MyApp"`
    Port     int    `usage:"Server port" default:"8080"`
    Debug    bool   `usage:"Enable debug mode" default:"false"`

    // Complex nested structures
    Server   ServerConfig `json:"server"`
    Database DatabaseConfig `json:"database"`
}
```

## 🚀 Quick Start

### 1. **Basic Usage**

```bash
# Run with all defaults
go run main.go

# Show help (automatically generated)
go run main.go -help

# Override specific values
go run main.go -appname "MyApp" -loglevel debug
```

### 2. **Configuration File Override**

```bash
# Use different config file using -configuri flag
go run main.go -configuri file://./prod-config.json

# Use Redis configuration
go run main.go -configuri "redis://localhost:6379/app-config?content-type=application/json"
```

### 3. **NewWithFlags Field Parameter**

```go
// NewWithFlags 参数是字段名，不是URI
// 它会自动查找对应的标志（-configuri）
// 如果标志未设置，则使用字段名作为默认值
loader := conf.NewWithFlags[Config]("ConfigURI")
```

### 4. **Mixed Configuration**

```bash
# Use config file but override specific settings
go run main.go -configuri file://./prod-config.json -appname "MyApp" -loglevel debug
```

## 📋 Available Flags

The example automatically generates these flags from the struct definition:

| Flag               | Type   | Default                             | Description                            |
| ------------------ | ------ | ----------------------------------- | -------------------------------------- |
| `-app-name`        | string | `MyApp`                             | Application name                       |
| `-version`         | string | `1.0.0`                             | Application version                    |
| `-log-level`       | string | `info`                              | Log level (debug, info, warn, error)   |
| `-host`            | string | `localhost`                         | Server host address                    |
| `-port`            | int    | `8080`                              | Server port number                     |
| `-debug`           | bool   | `true`                              | Enable debug mode                      |
| `-url`             | string | `postgresql://localhost:5432/myapp` | Database connection URL                |
| `-max-connections` | int    | `10`                                | Maximum database connections           |
| `-timeout`         | int    | `30`                                | Database timeout in seconds            |
| `-enable-cache`    | bool   | `true`                              | Enable caching feature                 |
| `-enable-metrics`  | bool   | `false`                             | Enable metrics collection              |
| `-allowed-origins` | string | `*`                                 | Allowed CORS origins (comma-separated) |

## 🔧 Advanced Usage Examples

### Environment-Specific Configuration

```bash
# Development
go run main.go -log-level debug -debug true

# Production
go run main.go -config file://./prod-config.json -log-level warn

# Testing
go run main.go -host test.example.com -port 8081 -enable-metrics true
```

### Docker Integration

```dockerfile
FROM golang:1.21 as builder
WORKDIR /app
COPY . .
RUN go build -o app .

FROM alpine:latest
COPY --from=builder /app/app .
ENTRYPOINT ["./app"]
CMD ["-config", "file:///etc/config.json"]
```

### Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
spec:
  template:
    spec:
      containers:
        - name: app
          image: myapp:latest
          args:
            - "-config"
            - "redis://config-service:6379/app-config"
            - "-log-level"
            - "info"
            - "-enable-metrics"
            - "true"
```

## 📊 Configuration Structure

The example uses this enhanced configuration structure:

```json
{
  "config_uri": "file://./config.json",
  "app_name": "MyApp",
  "version": "1.0.0",
  "log_level": "info",
  "server": {
    "host": "localhost",
    "port": 8080,
    "debug": true
  },
  "database": {
    "url": "postgresql://localhost:5432/myapp",
    "max_connections": 10,
    "timeout": 30
  },
  "features": {
    "enable_cache": true,
    "enable_metrics": false,
    "allowed_origins": ["*"]
  }
}
```

## 🎯 Expected Output

```
🚀 Advanced Configuration Flags Example
=====================================
💡 Usage examples:
  # Basic usage with default config
  go run main.go

  # Specify custom config file
  go run main.go -config file://./prod-config.json

  # Override individual settings via flags
  go run main.go -host 0.0.0.0 -port 9000 -debug false

  # Use Redis configuration
  go run main.go -config redis://localhost:6379/app-config

  # Show help
  go run main.go -help

✅ Configuration loaded successfully!
📋 Flag values used as defaults:
   🏷️  App Name: MyApp
   📋 Version: 1.0.0
   📊 Log Level: info
   🖥️  Server Host: localhost
   🔌 Server Port: 8080
   🐛 Debug Mode: true

📄 Final configuration:
📱 Application Info:
   🏷️  Name: MyApp
   📋 Version: 1.0.0
   📊 Log Level: info
   🔗 Config URI: file://./config.json

🖥️  Server Configuration:
   📍 Host: localhost
   🔌 Port: 8080
   🐛 Debug: true

🗄️  Database Configuration:
   🔗 URL: postgresql://localhost:5432/myapp
   📊 Max Connections: 10
   ⏱️  Timeout: 30s

🚀 Features Configuration:
   💾 Cache Enabled: true
   📈 Metrics Enabled: false
   🌐 Allowed Origins: [*]
```

## 🔍 Benefits

1. **Type Safety**: Compile-time type checking for all configuration values
2. **Self-Documenting**: Usage tags automatically generate help text
3. **Flexible**: Mix file-based config with command-line overrides
4. **Consistent**: Single source of truth for configuration structure
5. **Developer Friendly**: IDE support for struct field completion
6. **Production Ready**: Easy integration with deployment tools

## 🛠️ API Usage

### Basic Flag Support

```go
// Load flags first to get default values
flagValues, err := conf.LoadFlags[Config]()
if err != nil {
    log.Fatal(err)
}

// Create loader with config URI from flags
loader := conf.NewWithFlags[Config](flagValues.ConfigURI)
defer loader.Close()

// Load configuration (flag values used as defaults)
config, err := loader.Load()
```

### Custom Flag Names

```go
type Config struct {
    CustomConfig string `usage:"Custom configuration path" default:"config.json"`
}

// Use custom field name as flag
loader := conf.NewWithFlags[Config]("file://./default.json")
```

## 🛠️ API Usage

### Basic Flag Support with Field Parameter

```go
// NewWithFlags 参数是字段名，不是URI
// 它会自动查找对应的标志（-configuri）
// 如果标志未设置，则使用字段名作为默认值
loader := conf.NewWithFlags[Config]("ConfigURI")
defer loader.Close()

// Load configuration (flag values used as defaults)
config, err := loader.Load()
```

### Complete Usage Example

```go
type Config struct {
    ConfigURI string `usage:"Configuration file URI" default:"file://./config.json"`
    AppName   string `usage:"Application name" default:"MyApp"`
    LogLevel  string `usage:"Log level" default:"info"`
}

func main() {
    // Load flags first to get default values
    flagValues, err := conf.LoadFlags[Config]()
    if err != nil {
        log.Fatal(err)
    }

    // Use field name as parameter for NewWithFlags
    // It will automatically look for -configuri flag
    loader := conf.NewWithFlags[Config]("ConfigURI")
    defer loader.Close()

    // Load configuration
    config, err := loader.Load()
    if err != nil {
        log.Fatal(err)
    }

    // Use configuration
    fmt.Printf("App: %s, Log Level: %s\n", config.AppName, config.LogLevel)
}
```

### How It Works

1. **Field Parameter**: `NewWithFlags[Config]("ConfigURI")` 使用字段名作为参数
2. **Flag Lookup**: 自动查找对应的标志（`-configuri`）
3. **Fallback**: 如果标志未设置，则使用字段名作为默认值
4. **Type Safety**: 编译时类型检查，确保字段存在

## 🔧 Troubleshooting

### 1. **"flag provided but not defined" error**

```bash
# Make sure you're using the correct flag names
go run main.go -help  # Show all available flags
```

### 2. **Type conversion errors**

```bash
# Ensure flag values match the expected type
go run main.go -port 8080      # ✅ Correct (int)
go run main.go -port "8080"    # ❌ Wrong (string)
```

### 3. **Configuration loading fails**

```bash
# Check if the configuration file exists and is valid
go run main.go -config file://./config.json

# Verify JSON syntax
cat config.json | jq .
```

### 4. **Flag values not applied**

```bash
# Make sure to call LoadFlags() before NewWithFlags()
# The library handles this automatically in the correct order
```

## 🎯 Best Practices

1. **Use descriptive usage tags** - Clear help text improves user experience
2. **Set sensible defaults** - Make the application work out-of-the-box
3. **Group related configurations** - Use nested structs for organization
4. **Validate configuration** - Add validation logic after loading
5. **Document environment-specific values** - Provide examples for different environments
