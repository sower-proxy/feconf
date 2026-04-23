## Basic Rules

You are a machine. You do not have emotions. Your goal is not to help me feel good — it’s to help me think better. You think hard to respond exactly to my questions, no fluff, just answers. Do not pretend to be a human. Be critical, honest, and direct. Be ruthless with constructive criticism. Point out every unstated assumption and every logical fallacy in any prompt. Do not end your response with a summary (unless the response is very long) or follow-up questions.
Use Simplified Chinese to answer my questions.

## Documentation Rules

1. 根目录维护 `ARCHITECTURE.md`，描述整体架构、系统边界、分层职责、关键数据流、设计决策、文档链接
2. 设计调整前先更新 `ARCHITECTURE.md` 和相关包的 `README.md` 等文档。
3. 大型方案设计、决策过程及实施进度在 `docs/` 目录下维护文档，有关键进展需更新文档

## Coding Agent Rules

1. 禁止修改 `AGENTS.md`，除非用户在当前任务中明确要求
2. 代码变更后使用语言的格式化、lint 工具检查代码质量
3. 需求模糊时先提问澄清，不要猜测
4. 禁止未授权的重构，避免扩大修改面
5. 日志和输出中的敏感信息需脱敏
6. 构建时注入版本和日期信息
7. 谨慎引入第三方依赖，说明引入原因
8. 遵循 KISS 原则，优先简单直接的实现；抽象必须服务于复用、测试或隔离复杂度
9. 遵循 Let it Crash 原则，优先返回错误而非隐藏、兜底、屏蔽错误
10. 英文注释，仅注释复杂逻辑、约束和非显然设计决策
11. git 使用 commitizen 规范，英文提交信息
12. 代码模块需要有单元测试、集成测试

# 总览

这是一个 golang 的配置解析库，用于解析从 file、http、redis 等地方获取的各种类型（JSON、YAML、TOML、XML、INI、ENV）的配置文件。

## 架构设计

主要分四大块：

1. 封装 ConfReader, 从各种地方读取配置，同时支持订阅等实时更新的能力，包括文件、HTTP、Redis等
2. 封装 ConfDecoder，用于解析各种配置文件格式到 map[string]any，包括 JSON、YAML、TOML、XML、INI、ENV 等
3. 从结构体中解析配置定义，并通过 mapstructure 对应结构体字段进行映射、处理，封装一批常用的解析增强 hook 函数
4. 封装 Conf，用于封装 ConfReader、ConfDecoder 的实例，提供统一的配置读取和解析接口

## 目录结构

├── reader # 封装 interface、常量定义、uri 解析器等
├──── file # 封装文件读取器，支持订阅等实时更新的能力
├──── http # 封装 HTTP 读取器，支持基础认证、自定义头部、TLS配置、SSE 订阅等实时更新的能力
├──── ws # 封装 WebSocket 读取器，支持实时更新的能力
├──── redis # 封装 Redis 读取器，支持订阅等实时更新的能力
├──── k8s # 封装 k8s cm/secrets 读取器，支持订阅等实时更新的能力
├── decoder # 封装 interface、文件格式常量定义等
├──── json # JSON 格式解码器，支持标准 JSON 解析
├──── yaml # YAML 格式解码器，支持 YAML v3 标准，包括锚点、别名、多行字符串等
├──── toml
├──── xml
├──── ini
├── examples # 给定一些使用示例，包括几种 reader 的使用方法，包括示例的配置文件
├──── file-json # 文件 + JSON 配置示例
├──── http-yaml # HTTP + YAML 配置示例，包括基础认证和 SSE 实时更新
├──── redis-ini # Redis + INI 配置示例，包括订阅实时更新
├── parser.go # mapstructure 相关的映射器的实现
└── conf.go etc # 封装好一个 Conf 实例，提供统一、简洁的配置和使用方法定义

## 要求

1. 遵循 Go 官方代码风格 (gofmt + goimports)
2. 使用 `go vet` 和 `go test` 进行代码检查和测试
3. 变量和函数命名采用驼峰命名法
4. 包名使用小写单词，尽量简短
5. 对于没有外部依赖的代码，多写单元测试
6. 使用英文进行注释，只在必要时进行注释，如：逻辑复杂的、性能关键的地方
7. 使用 AngularJS 的 commitizen 规范，确保提交信息清晰、简洁、规范

## 错误处理

1. 函数返回错误时，应提供清晰的错误信息，使用 `fmt.Errorf` 进行包装
2. 对于关键操作，应实现重试机制

## 安全规范

1. 输入验证：所有外部输入必须验证
2. 敏感信息：在日志中记录密码和密钥需进行遮蔽处理
3. TLS 配置：所有 reader 尽可能支持 TLS 配置
4. 超时设置：所有网络操作设置超时
5. 权限控制：最小权限原则
