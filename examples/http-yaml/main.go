package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sower-proxy/feconf"
	_ "github.com/sower-proxy/feconf/decoder/yaml"
	_ "github.com/sower-proxy/feconf/reader/http"
)

func main() {
	fmt.Println("=== HTTP + YAML 配置示例 ===")

	// 启动配置服务器（后台运行）
	fmt.Println("启动配置服务器...")
	go runServer()
	time.Sleep(2 * time.Second) // 等待服务器启动

	// 运行基础示例
	fmt.Println("运行基础配置示例...")
	runBasicExample()

	// 运行认证示例
	fmt.Println("运行认证订阅示例... (需要约20秒)")
	runAuthExample()

	fmt.Println("✅ 所有示例完成")
}

// 基础配置加载示例
func runBasicExample() {
	configURI := "http://localhost:8080/config.yaml"
	loader := feconf.New[Config](configURI)
	defer loader.Close()

	var config Config
	err := loader.Load(&config)
	if err != nil {
		fmt.Printf("❌ 基础示例失败: %v\n", err)
		return
	}

	fmt.Println("✅ 基础配置加载成功!")
	fmt.Printf("  关键配置信息:\n")
	fmt.Printf("    应用: %s v%s (%s)\n",
		config.App.Name, config.App.Version, config.App.Environment)
	fmt.Printf("    服务器: %s:%d (调试: %t)\n",
		config.Server.Host, config.Server.Port, config.Server.Debug)
	fmt.Printf("    数据库: %s\n", config.Database.Primary.URL)
	fmt.Printf("    从库数量: %d\n", len(config.Database.Replicas))
	fmt.Printf("    功能特性: %v\n", config.Features)
	fmt.Println()
}

// 认证配置订阅示例
func runAuthExample() {
	configURI := "http://user:pass@localhost:8080/config-auth.yaml"
	loader := feconf.New[Config](configURI)
	defer loader.Close()

	fmt.Println("🔐 启动认证配置订阅...")

	// 订阅配置变更
	eventChan, err := loader.Subscribe()
	if err != nil {
		fmt.Printf("❌ 认证订阅失败: %v\n", err)
		fmt.Println("提示: 确保服务器支持认证 (user:pass)")
		return
	}

	fmt.Println("✅ 认证订阅成功!")
	fmt.Println("监听配置变更中... (预期收到 3-4 个事件)")
	fmt.Println("  - 初始配置事件")
	fmt.Println("  - 3个更新事件 (每4秒一个)")
	fmt.Println()
	// 创建超时上下文 - 监听20秒，确保能收到所有更新
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	eventCount := 0
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("⏰ 认证订阅完成 (收到 %d 个配置事件)\n", eventCount)
			return

		case event, ok := <-eventChan:
			if !ok {
				fmt.Println("📞 认证事件通道关闭")
				return
			}

			eventCount++
			if event.IsValid() {
				fmt.Printf("📝 [事件 %d] 认证配置更新成功\n", eventCount)
				fmt.Printf("  时间: %s\n", event.Timestamp.Format("15:04:05"))
				fmt.Printf("  来源: %s\n", event.SourceURI)
				fmt.Printf("  关键配置变更:\n")
				fmt.Printf("    应用: %s v%s (%s)\n",
					event.Config.App.Name, event.Config.App.Version, event.Config.App.Environment)
				fmt.Printf("    服务器: %s:%d (调试: %t)\n",
					event.Config.Server.Host, event.Config.Server.Port, event.Config.Server.Debug)
				fmt.Printf("    数据库: %s\n", event.Config.Database.Primary.URL)
				fmt.Printf("    从库数量: %d\n", len(event.Config.Database.Replicas))
				fmt.Printf("    功能特性: %v\n", event.Config.Features)
				fmt.Println("  " + strings.Repeat("─", 50))
			} else {
				fmt.Printf("❌ [事件 %d] 认证配置更新失败: %v\n", eventCount, event.Error)
			}
		}
	}
}
