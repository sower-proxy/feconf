package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

var configData []byte

// runServer 运行简化的配置服务器
func runServer() {
	// 读取配置文件
	data, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Printf("读取配置文件失败: %v", err)
		// 创建默认配置
		configData = []byte(`app:
  name: "示例应用"
  version: "1.0.0"
  environment: "development"
server:
  host: "localhost"
  port: 8080
  debug: true
database:
  primary:
    host: "localhost"
    port: 5432
    name: "myapp"
    url: "postgres://user:pass@localhost:5432/myapp"
  replicas: []
features: ["auth", "logging", "metrics"]`)
	} else {
		configData = data
	}

	// 设置路由
	http.HandleFunc("/config.yaml", handleConfig)
	http.HandleFunc("/config-auth.yaml", handleConfigAuth)
	http.HandleFunc("/", handleIndex)

	// 启动模拟配置更新协程
	go simulateConfigUpdates()

	fmt.Println("配置服务器运行在 http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, `<h1>配置服务器</h1>
<p><a href="/config.yaml">基础配置</a></p>
<p><a href="/config-auth.yaml">认证配置 (user:pass)</a></p>
<p>注意: 配置会自动更新以演示订阅功能</p>`)
}

func handleConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/yaml")
	w.Write(configData)
}

func handleConfigAuth(w http.ResponseWriter, r *http.Request) {
	username, password, ok := r.BasicAuth()
	if !ok || username != "user" || password != "pass" {
		w.Header().Set("WWW-Authenticate", `Basic realm="Config"`)
		http.Error(w, "认证失败", http.StatusUnauthorized)
		return
	}

	// 检查是否是 SSE 请求
	if r.Header.Get("Accept") == "text/event-stream" {
		handleConfigAuthSSE(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/yaml")
	w.Write(configData)
}

func handleConfigAuthSSE(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// 发送初始配置
	fmt.Fprintf(w, "data: %s\n\n", string(configData))
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}

	// 等待一下再开始发送更新
	time.Sleep(2 * time.Second)

	// 模拟定期配置更新
	ticker := time.NewTicker(4 * time.Second)
	defer ticker.Stop()

	for i := 1; i <= 3; i++ { // 发送 3 次更新
		select {
		case <-ticker.C:
			// 模拟配置更新
			updatedConfig := fmt.Sprintf(`app:
  name: "示例应用"
  version: "1.0.%d"
  environment: "development"
server:
  host: "localhost"
  port: 8080
  debug: true
database:
  primary:
    host: "localhost"
    port: 5432
    name: "myapp"
    url: "postgres://user:pass@localhost:5432/myapp"
  replicas: []
features: ["auth", "logging", "metrics", "update-%d"]`, i, i)

			fmt.Fprintf(w, "data: %s\n\n", updatedConfig)
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}
			fmt.Printf("📝 服务器: SSE 配置更新 %d 已发送\n", i)
		case <-r.Context().Done():
			return
		}
	}
}

// 模拟配置文件更新
func simulateConfigUpdates() {
	time.Sleep(8 * time.Second) // 等待订阅开始

	for i := 1; i <= 2; i++ {
		time.Sleep(6 * time.Second)

		// 更新内存中的配置数据
		configData = fmt.Appendf(nil, `# 自动更新的配置 %d
app:
  name: "示例应用"
  version: "1.0.%d"
  environment: "development"

server:
  host: "localhost"
  port: 8080
  debug: true

database:
  primary:
    host: "localhost"
    port: 5432
    name: "myapp"
    url: "postgres://user:pass@localhost:5432/myapp"
  replicas:
    - "postgres://user:pass@slave1:5432/myapp"

features:
  - "auth"
  - "logging"
  - "metrics"
  - "auto-update-%d"`, i, i, i)

		fmt.Printf("📝 服务器: 配置已更新 (版本 1.0.%d)\n", i)
	}
}
