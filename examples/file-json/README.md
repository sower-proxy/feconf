# File + JSON 配置示例

演示如何使用配置库加载 JSON 配置文件。

## 文件

- `config.json` - 示例 JSON 配置文件，支持环境变量替换
- `main.go` - 简单的配置加载示例

## 运行示例

```bash
cd examples/file-json
go run main.go
```

注意：示例中使用绝对路径，实际使用时可根据需要调整路径。

## 环境变量支持

配置文件支持环境变量替换：

```bash
export HOST=127.0.0.1
export PORT=9000
export DEBUG=true
go run main.go
```

## 输出示例

```
Server: localhost:8080 (debug: false)
Database: sqlite:///app.db (max_conn: 10)
```