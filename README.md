# conf

[![Go Version](https://img.shields.io/badge/go-%3E%3D1.23-blue.svg)](https://golang.org/)

A flexible, URI-based configuration management library for Go that supports multiple protocols, formats, and real-time configuration updates.

## Features

- 🌐 **Multiple Protocols**: Support for HTTP/HTTPS, WebSocket, Redis, and local files
- 📄 **Multiple Formats**: JSON, YAML, INI, TOML, and XML configuration formats
- 🔄 **Real-time Updates**: Subscribe to configuration changes with automatic reloading
- 🚩 **Command-line Flags**: Built-in support for reading configuration URI from flags
- 🔌 **Extensible**: Plugin-based architecture for custom readers and decoders
- ⚡ **Performance**: Efficient parsing with connection pooling and retry mechanisms
- 🔒 **Security**: TLS support with custom certificates and authentication
- 📋 **Type Safety**: Strong typing with struct mapping using mapstructure

## Installation

```bash
go get github.com/sower-proxy/conf
```

## Quick Start

```go
package main

import (
    "fmt"
    "log"

    "github.com/sower-proxy/conf"
    _ "github.com/sower-proxy/conf/decoder/json"
    _ "github.com/sower-proxy/conf/reader/file"
)

type Config struct {
    Host string `json:"host"`
    Port int    `json:"port"`
}

func main() {
    loader := conf.New[Config]("file://./config.json")
    defer loader.Close()

    config, err := loader.Load()
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Server: %s:%d\n", config.Host, config.Port)
}
```

For command-line flag support:
```go
// Read config URI from -config flag, fallback to default if not provided
loader := conf.NewWithFlags[Config]("file://./default-config.json")

// Or use custom flag name
loader := conf.NewWithFlagsNamed[Config]("app-config", "file://./default-config.json")
```

For real-time updates:
```go
eventChan, _ := loader.Subscribe()
for event := range eventChan {
    if event.IsValid() {
        fmt.Printf("Config updated: %v\n", event.Config)
    }
}
```

## Supported Protocols

### File System
```go
loader := conf.New[Config]("file:///path/to/config.json")
```

### HTTP/HTTPS
```go
loader := conf.New[Config]("https://api.example.com/config.yaml?timeout=30s")
```

### WebSocket
```go
loader := conf.New[Config]("wss://realtime.example.com/config?ping_interval=30s")
```

### Redis
```go
loader := conf.New[Config]("redis://localhost:6379/config-key?db=1&pool_size=10")

// For Redis hash fields
loader := conf.New[Config]("redis://localhost:6379/config-hash#field-name")
```

### Kubernetes ConfigMap/Secret
```go
// Read from a ConfigMap (returns first key's value)
loader := conf.New[Config]("k8s://configmap/default/app-config")

// Read a specific key from a ConfigMap (e.g., a configuration file)
loader := conf.New[Config]("k8s://configmap/default/app-config/config.yaml")

// Read from a Secret (returns first key's value)
loader := conf.New[string]("k8s://secret/default/db-secret")

// Read a specific key from a Secret (e.g., a password)
loader := conf.New[string]("k8s://secret/default/db-secret/password")
```

The Kubernetes reader supports reading configuration data stored as keys in ConfigMaps or Secrets. When a specific key is provided in the URI, only that key's value is read. When no key is specified, the reader returns the first key's value if there's only one key, or returns an error if there are multiple keys.

## Supported Formats

The library automatically detects format from file extensions or can be specified via `content-type` parameter:

- **JSON**: `.json` or `content-type=application/json`
- **YAML**: `.yaml`, `.yml` or `content-type=application/yaml`
- **INI**: `.ini` or `content-type=text/ini`
- **TOML**: `.toml` or `content-type=application/toml`

- **XML**: `.xml` or `content-type=application/xml`

## Examples

The `examples/` directory contains complete working examples:

- **File + JSON**: Basic file-based configuration with JSON format
- **HTTP + YAML**: HTTP-based configuration with YAML format and authentication
- **Redis + INI**: Redis-based configuration with INI format
- **Redis + Hash**: Redis hash field configuration with JSON format
- **WebSocket + XML**: Real-time configuration updates via WebSocket with XML format
<<<<<<< Updated upstream
=======
- **Flags**: Command-line flag support for configuration URI specification

>>>>>>> Stashed changes
### Running Examples

```bash
# File example
cd examples/file-json && go run .

# HTTP example (starts a demo server)
cd examples/http-yaml && go run .

# Redis example (requires Redis server)
cd examples/redis-ini && ./setup.sh && go run .

# Redis hash fields example (requires Redis server)
cd examples/redis-hash && ./setup.sh && go run .

# Command-line flags example
cd examples/flags && go run . -config file://./prod-config.json

# WebSocket example (starts a demo WebSocket server)
cd examples/ws-xml && go run .
```

## Architecture

The library follows a modular plugin-based architecture:

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Application   │    │   Configuration  │    │   Decoder       │
│                 │───▶│   Loader (conf)  │───▶│   (json/yaml/   │
│   Your Code     │    │                  │    │    ini/etc)     │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                                │
                                ▼
                       ┌──────────────────┐
                       │   Reader         │
                       │   (file/http/    │
                       │   redis/ws)      │
                       └──────────────────┘
```

### Components

- **Configuration Loader**: Main entry point that coordinates readers and decoders
- **Readers**: Protocol-specific implementations (file, HTTP, Redis, WebSocket)
- **Decoders**: Format-specific parsers (JSON, YAML, INI, TOML, XML)
- **Event System**: Real-time configuration change notifications

## Best Practices

1. **Import required modules**: Always import specific decoder and reader modules
2. **Handle errors**: Always check and handle configuration loading errors
3. **Resource cleanup**: Use `defer loader.Close()` to cleanup resources
4. **Use timeouts**: Leverage context for request timeouts
5. **Validate configuration**: Validate loaded configuration before use

## Advanced Features

### Custom Mapstructure Configuration
```go
loader := conf.New[Config](uri)
loader.ParserConf.TagName = "config"  // Use 'config' tags instead of 'json'
loader.ParserConf.ErrorUnused = true  // Error on unused fields
```

### Environment-specific Configuration
```go
env := os.Getenv("APP_ENV")
if env == "" {
    env = "development"
}

uri := fmt.Sprintf("file://./config-%s.yaml", env)
loader := conf.New[Config](uri)
```

### Configuration Hot-reloading with Graceful Shutdown
```go
eventChan, err := loader.Subscribe()
if err != nil {
    return err
}

go func() {
    for event := range eventChan {
        if event.IsValid() {
            // Atomically update configuration
            atomic.StorePointer(&currentConfig, unsafe.Pointer(event.Config))
            log.Println("Configuration hot-reloaded successfully")
        }
    }
}()
```



## URI Query Parameters

The configuration library supports various special query parameters during URI parsing to customize the behavior of different readers and decoders.

### Universal Parameters

| Parameter | Type | Default | Description | Example |
|-----------|------|---------|-------------|---------|
| `content-type` | string | From file extension | Specifies the MIME type to determine decoder format | `?content-type=application/json` |

### HTTP/HTTPS Parameters

| Parameter | Type | Default | Validation | Description | Example |
|-----------|------|---------|------------|-------------|---------|
| `timeout` | duration | `30s` | - | HTTP request timeout duration | `?timeout=60s` |
| `retry_attempts` | integer | `3` | ≥ 1 | Number of retry attempts for failed requests | `?retry_attempts=5` |
| `retry_delay` | duration | `1s` | - | Delay between retry attempts | `?retry_delay=2s` |
| `header_*` | string | - | - | Custom HTTP headers (format: `header_<name>=<value>`) | `?header_Authorization=Bearer%20token` |
| `tls_insecure` | boolean | `false` | - | Skip TLS certificate verification | `?tls_insecure=true` |

### WebSocket (WS/WSS) Parameters

| Parameter | Type | Default | Validation | Description | Example |
|-----------|------|---------|------------|-------------|---------|
| `timeout` | duration | `30s` | - | WebSocket handshake timeout | `?timeout=45s` |
| `retry_attempts` | integer | `3` | ≥ 1 | Number of connection retry attempts | `?retry_attempts=10` |
| `retry_delay` | duration | `1s` | - | Delay between connection retry attempts | `?retry_delay=5s` |
| `ping_interval` | duration | `30s` | - | Interval for sending ping messages | `?ping_interval=60s` |
| `pong_wait` | duration | `60s` | - | Maximum wait time for pong response | `?pong_wait=120s` |
| `write_wait` | duration | `10s` | - | Timeout for WebSocket write operations | `?write_wait=15s` |
| `header_*` | string | - | - | Custom WebSocket headers (format: `header_<name>=<value>`) | `?header_Origin=https://example.com` |
| `tls_insecure` | boolean | `false` | - | Skip TLS certificate verification for WSS | `?tls_insecure=true` |

### Redis Parameters

| Parameter | Type | Default | Validation | Description | Example |
|-----------|------|---------|------------|-------------|---------|
| `db` | integer | `0` | ≥ 0 | Redis database number | `?db=1` |
| `timeout` | duration | `30s` | - | Redis operation timeout | `?timeout=10s` |
| `retry_delay` | duration | `1s` | - | Delay between retry attempts | `?retry_delay=500ms` |
| `max_retries` | integer | `3` | ≥ 0 | Maximum number of retry attempts | `?max_retries=5` |
| `pool_size` | integer | `10` | > 0 | Redis connection pool size | `?pool_size=20` |
| `min_idle_conns` | integer | `1` | ≥ 0 | Minimum idle connections in pool | `?min_idle_conns=5` |
| `tls_insecure` | boolean | `false` | - | Skip TLS certificate verification for REDISS | `?tls_insecure=true` |

**Hash Field Support**: Use URI fragment (`#field-name`) to read from specific hash fields:
```
redis://localhost:6379/user:123?content-type=application/json#profile
redis://localhost:6379/settings?content-type=application/json#database
```
Note: When using hash fields, you must specify `content-type` parameter for format detection.

## URI Examples

### HTTP with Custom Parameters
```
http://config-server.example.com/app-config.json?timeout=60s&retry_attempts=5&header_Authorization=Bearer%20mytoken
```

### WebSocket with Protocol Configuration
```
wss://realtime-config.example.com/config?ping_interval=30s&pong_wait=90s&tls_insecure=false
```

### Redis with Database and Pool Settings
```
redis://localhost:6379/my-config-key?db=2&pool_size=15&max_retries=3&timeout=10s
```

### Redis with Hash Field
```
redis://localhost:6379/app-settings?db=1&timeout=5s&content-type=application/json#database
```

### File with Content Type Override
```
file:///path/to/config.txt?content-type=application/yaml
```


## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for version history and migration guides.
