# feconf

[![Go Version](https://img.shields.io/badge/go-%3E%3D1.24-blue.svg)](https://golang.org/)
[![Go Report Card](https://goreportcard.com/badge/github.com/sower-proxy/feconf)](https://goreportcard.com/report/github.com/sower-proxy/feconf)
[![License: MPL 2.0](https://img.shields.io/badge/License-MPL%202.0-brightgreen.svg)](https://opensource.org/licenses/MPL-2.0)
[![Test Coverage](https://img.shields.io/badge/coverage-92%25-brightgreen.svg)](https://github.com/sower-proxy/feconf/actions/workflows/test.yml)
[![Build Status](https://img.shields.io/github/actions/workflow/status/sower-proxy/feconf/test.yml?branch=main)](https://github.com/sower-proxy/feconf/actions/workflows/test.yml)
[![Documentation](https://img.shields.io/badge/docs-godoc-blue.svg)](https://pkg.go.dev/github.com/sower-proxy/feconf)
[![Code Size](https://img.shields.io/github/languages/code-size/sower-proxy/feconf)](https://github.com/sower-proxy/feconf)

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
- 🎣 **Extensible Hooks**: Built-in hook functions for type conversion and data processing

## Installation

```bash
go get github.com/sower-proxy/feconf
```

## Quick Start

```go
package main

import (
    "fmt"
    "log"

    "github.com/sower-proxy/feconf"
    _ "github.com/sower-proxy/feconf/decoder/json"
    _ "github.com/sower-proxy/feconf/reader/file"
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
// Read from a ConfigMap
loader := conf.New[Config]("k8s://configmap/default/app-config")

// Read a specific key from a ConfigMap
loader := conf.New[Config]("k8s://configmap/default/app-config/config.yaml")

// Read from a Secret
loader := conf.New[string]("k8s://secret/default/db-secret/password")
```

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
- **Kubernetes**: Kubernetes ConfigMap/Secret configuration with YAML format
- **Flags**: Command-line flag support for configuration URI specification

### Running Examples

```bash
# File example
cd examples/file-json && go run .

# HTTP example (starts a demo server)
cd examples/http-yaml && go run .

# Redis example (requires Redis server)
cd examples/redis-ini && ./setup.sh && go run .

# Command-line flags example
cd examples/flags && go run . -config file://./prod-config.json

# WebSocket example (starts a demo WebSocket server)
cd examples/ws-xml && go run .

# Kubernetes example (requires Kubernetes cluster)
cd examples/k8s-yaml && ./setup.sh && go run .
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

## Best Practices

1. **Import required modules**: Always import specific decoder and reader modules
2. **Handle errors**: Always check and handle configuration loading errors
3. **Resource cleanup**: Use `defer loader.Close()` to cleanup resources
4. **Use timeouts**: Leverage context for request timeouts
5. **Validate configuration**: Validate loaded configuration before use
6. **Use appropriate hook functions**: Leverage built-in hook functions for common data transformations

## Advanced Features

### Custom Mapstructure Configuration

```go
loader := conf.New[Config](uri)
loader.ParserConf.TagName = "config"  // Use 'config' tags instead of 'json'
loader.ParserConf.ErrorUnused = true  // Error on unused fields

// Custom hook functions can be added to the decode chain
customHook := mapstructure.ComposeDecodeHookFunc(
    conf.HookFuncDefault(),      // Default value handling
    conf.HookFuncEnvRender(),     // Environment variable rendering
    conf.HookFuncStringToBool(),  // String to boolean conversion
    conf.HookFuncStringToSlogLevel(), // String to log level conversion
    // Add your custom hooks here
)
loader.ParserConf.DecodeHook = customHook
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

## URI Query Parameters

### Universal Parameters

| Parameter      | Type   | Default             | Description                                         | Example                          |
| -------------- | ------ | ------------------- | --------------------------------------------------- | -------------------------------- |
| `content-type` | string | From file extension | Specifies the MIME type to determine decoder format | `?content-type=application/json` |

### HTTP/HTTPS Parameters

| Parameter        | Type     | Default | Description                                           | Example                                |
| ---------------- | -------- | ------- | ----------------------------------------------------- | -------------------------------------- |
| `timeout`        | duration | `30s`   | HTTP request timeout duration                         | `?timeout=60s`                         |
| `retry_attempts` | integer  | `3`     | Number of retry attempts for failed requests          | `?retry_attempts=5`                    |
| `retry_delay`    | duration | `1s`    | Delay between retry attempts                          | `?retry_delay=2s`                      |
| `header_*`       | string   | -       | Custom HTTP headers (format: `header_<name>=<value>`) | `?header_Authorization=Bearer%20token` |
| `tls_insecure`   | boolean  | `false` | Skip TLS certificate verification                     | `?tls_insecure=true`                   |

### Redis Parameters

| Parameter      | Type     | Default | Description                                  | Example              |
| -------------- | -------- | ------- | -------------------------------------------- | -------------------- |
| `db`           | integer  | `0`     | Redis database number                        | `?db=1`              |
| `timeout`      | duration | `30s`   | Redis operation timeout                      | `?timeout=10s`       |
| `pool_size`    | integer  | `10`    | Redis connection pool size                   | `?pool_size=20`      |
| `tls_insecure` | boolean  | `false` | Skip TLS certificate verification for REDISS | `?tls_insecure=true` |

**Hash Field Support**: Use URI fragment (`#field-name`) to read from specific hash fields:

```
redis://localhost:6379/user:123?content-type=application/json#profile
```

## License

This project is licensed under the Mozilla Public License Version 2.0 - see the [LICENSE](LICENSE) file for details.
