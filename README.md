# feconf

[![Go Version](https://img.shields.io/badge/go-%3E%3D1.24-blue.svg)](https://golang.org/)
[![License: MPL 2.0](https://img.shields.io/badge/License-MPL%202.0-brightgreen.svg)](https://opensource.org/licenses/MPL-2.0)
[![Build Status](https://img.shields.io/github/actions/workflow/status/sower-proxy/feconf/test.yml?branch=main)](https://github.com/sower-proxy/feconf/actions/workflows/test.yml)
[![Test Coverage](https://img.shields.io/badge/coverage-92%25-brightgreen.svg)](https://github.com/sower-proxy/feconf/actions/workflows/test.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/sower-proxy/feconf)](https://goreportcard.com/report/github.com/sower-proxy/feconf)
[![Documentation](https://img.shields.io/badge/docs-godoc-blue.svg)](https://pkg.go.dev/github.com/sower-proxy/feconf)

A flexible and comfortable, URI-based configuration management library for Go with real-time updates.

## Features

- **Multi-protocol**: File, HTTP, WebSocket, Redis, Kubernetes
- **Multi-format**: JSON, YAML, INI, TOML, XML
- **Real-time updates**: Subscribe to configuration changes
- **Type-safe**: Strong struct mapping with mapstructure
- **Extensible**: Plugin-based architecture
- **Production-ready**: TLS, retries, connection pooling

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
    loader := feconf.New[Config]("file://./config.json")
    defer loader.Close()

    config, err := loader.Load()
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Server: %s:%d\n", config.Host, config.Port)
}
```

## Supported Protocols

```go
// File system
loader := feconf.New[Config]("file:///path/to/config.yaml")

// HTTP/HTTPS
loader := feconf.New[Config]("https://api.example.com/config.json")

// Redis
loader := feconf.New[Config]("redis://localhost:6379/config-key")

// WebSocket
loader := feconf.New[Config]("wss://realtime.example.com/config")

// Kubernetes
loader := feconf.New[Config]("k8s://configmap/default/app-config")
```

## Supported Formats

- **JSON**: `.json` or `content-type=application/json`
- **YAML**: `.yaml`, `.yml` or `content-type=application/yaml`
- **INI**: `.ini` or `content-type=text/ini`
- **TOML**: `.toml` or `content-type=application/toml`
- **XML**: `.xml` or `content-type=application/xml`

## Real-time Updates

```go
eventChan, _ := loader.Subscribe()
for event := range eventChan {
    if event.IsValid() {
        fmt.Printf("Config updated: %v\n", event.Config)
    }
}
```

## Command-line Flags

```go
// Read from -config flag with fallback
loader := feconf.NewWithFlags[Config]("file://./default-config.json")
```

## Architecture

The library follows a modular plugin-based architecture:

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Application   │    │   Configuration  │    │   Decoder       │
│                 │───▶│   Loader (feconf)│───▶│   (json/yaml/   │
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

## Examples

See the `examples/` directory for complete working examples:

- `file-json/` - Basic file-based configuration
- `http-yaml/` - HTTP with authentication
- `redis-ini/` - Redis configuration
- `flags/` - Command-line flag integration
- `ws-xml/` - WebSocket real-time updates
- `k8s-yaml/` - Kubernetes ConfigMap/Secret

Run examples:

```bash
cd examples/file-json && go run .
cd examples/http-yaml && go run .
```

## Advanced Configuration

```go
loader := feconf.New[Config](uri)

// Custom parser settings
loader.ParserConf.TagName = "config"
loader.ParserConf.ErrorUnused = true

// Custom decode hooks
loader.ParserConf.DecodeHook = mapstructure.ComposeDecodeHookFunc(
    feconf.HookFuncDefault(),
    feconf.HookFuncEnvRender(),
    feconf.HookFuncStringToBool(),
)
```

## URI Parameters

Common parameters:

- `content-type` - Override format detection
- `timeout` - Request/operation timeout
- `retry_attempts` - Number of retries

Protocol-specific parameters available for HTTP, Redis, and WebSocket connections.

## Installation

```bash
go get github.com/sower-proxy/feconf
```

## License

Mozilla Public License 2.0 - See [LICENSE](LICENSE) for details.
