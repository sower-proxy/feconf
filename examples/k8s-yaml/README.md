# Kubernetes YAML 配置示例

演示如何使用配置库加载 Kubernetes 风格的 YAML 配置文件。

## 文件

- `config.yaml` - 示例 YAML 配置文件，支持环境变量替换
- `main.go` - 简单的配置加载示例
- `go.mod` - Go 模块定义文件

## 运行示例

首次运行需要先安装依赖：

```bash
cd examples/k8s-yaml
go get github.com/sower-proxy/conf@v0.0.0-00010101000000-000000000000
go get github.com/sower-proxy/conf/decoder/yaml@v0.0.0-00010101000000-000000000000
go get github.com/sower-proxy/conf/reader/file@v0.0.0-00010101000000-000000000000
go run main.go
```

或者使用简化命令：

```bash
cd examples/k8s-yaml
go run main.go
```

如果遇到依赖问题，可以尝试：

```bash
cd examples/k8s-yaml
go mod tidy
```

注意：示例中使用相对路径，实际使用时可根据需要调整路径。

## 环境变量支持

配置文件支持环境变量替换：

```bash
export HOST=0.0.0.0
export PORT=9000
export DEBUG=true
go run main.go
```

## 输出示例

```
Server: 0.0.0.0:8080 (debug: false)
Database: sqlite:///app.db (max_conn: 20)
Redis: localhost:6379 (db: 0)

=== Starting subscription to watch for config changes ===
Watching for configuration changes... (modify config.yaml to see updates)
Press Ctrl+C to exit or wait for timeout
[2025-09-02 13:56:28] Config updated from: file://./config.yaml
  Server: 0.0.0.0:8080 (debug: false)
  Database: sqlite:///app.db (max_conn: 20)
  Redis: localhost:6379 (db: 0)
Subscription timeout, exiting...
```