# HTTP + YAML 配置示例

演示如何使用配置库通过 HTTP 加载远程 YAML 配置文件，支持基础认证和实时配置订阅。

## 文件

- `main.go` - 统一的示例入口点，包含基础加载和认证订阅示例
- `server.go` - 简化的配置服务器，支持 SSE 配置推送
- `types.go` - 配置结构体定义
- `config.yaml` - 示例配置文件

## 运行示例

```bash
cd examples/http-yaml
go run .
```

程序会自动：
1. 启动配置服务器 (后台)
2. 运行基础配置加载示例 (一次性加载)
3. 运行认证配置订阅示例 (实时监听配置变更)

## 功能演示

- **基础加载**: 从 HTTP 服务器一次性加载配置
- **认证订阅**: 使用 HTTP Basic Authentication 订阅配置变更
- **实时更新**: 通过 Server-Sent Events (SSE) 接收配置更新
- **自动模拟**: 服务器自动模拟配置更新以演示订阅功能

## 示例输出

```
=== HTTP + YAML 配置示例 ===
启动配置服务器...
运行基础配置示例...
✅ 基础配置加载成功!
  关键配置信息:
    应用: 示例应用 v1.0.0 (development)
    服务器: localhost:8080 (调试: true)
    数据库: postgres://user:pass@localhost:5432/myapp
    从库数量: 1
    功能特性: [auth logging metrics]

运行认证订阅示例... (需要约20秒)
🔐 启动认证配置订阅...
✅ 认证订阅成功!
监听配置变更中... (预期收到 3-4 个事件)

📝 [事件 1] 认证配置更新成功
  时间: 14:30:15
  来源: http://user:pass@localhost:8080/config-auth.yaml
  关键配置变更:
    应用: 示例应用 v1.0.0 (development)
    服务器: localhost:8080 (调试: true)
    数据库: postgres://user:pass@localhost:5432/myapp
    从库数量: 0
    功能特性: [auth logging metrics]
  ──────────────────────────────────────────────────

📝 [事件 2] 认证配置更新成功
  时间: 14:30:19
  关键配置变更:
    应用: 示例应用 v1.0.1 (development)
    功能特性: [auth logging metrics update-1]
  ──────────────────────────────────────────────────

⏰ 认证订阅完成 (收到 4 个配置事件)
✅ 所有示例完成
```

## 服务端点

- `GET /` - 服务器首页
- `GET /config.yaml` - 基础配置 (无认证)
- `GET /config-auth.yaml` - 认证配置 (user:pass)
  - 支持 SSE: `Accept: text/event-stream`