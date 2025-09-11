package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sower-proxy/feconf"
	_ "github.com/sower-proxy/feconf/decoder/xml"
	_ "github.com/sower-proxy/feconf/reader/ws"
)

type Config struct {
	Server struct {
		Host string `xml:"host"`
		Port int    `xml:"port"`
	} `xml:"server"`
	Database struct {
		URL string `xml:"url"`
	} `xml:"database"`
}

func main() {
	fmt.Println("=== WebSocket + XML 配置示例 ===")

	// Generate TLS certificate and start secure WebSocket server
	fmt.Println("🔐 生成 TLS 证书...")
	if err := generateTLSCert(); err != nil {
		fmt.Printf("❌ 证书生成失败: %v\n", err)
		return
	}

	fmt.Println("🚀 启动 WebSocket 服务器...")
	go startWSServer()
	time.Sleep(2 * time.Second) // Wait for server to start

	// Load configuration via regular ws
	configURI := "ws://localhost:8080/config.xml"
	loader := feconf.New[Config](configURI)
	defer loader.Close()

	fmt.Println("🔄 加载配置...")
	var config Config
	err := loader.Load(&config)
	if err != nil {
		fmt.Printf("❌ 配置加载失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 配置加载成功: %s:%d, DB: %s\n",
		config.Server.Host, config.Server.Port, config.Database.URL)

	// Subscribe via secure wss
	secureURI := "wss://localhost:8443/config.xml?tls_insecure=true"
	secureLoader := feconf.New[Config](secureURI)
	defer secureLoader.Close()

	fmt.Println("📡 订阅配置更新 (WSS)...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	eventChan, err := secureLoader.Subscribe()
	if err != nil {
		fmt.Printf("❌ 订阅失败: %v\n", err)
		return
	}

	fmt.Println("✅ 订阅成功，监听配置变更...")
	eventCount := 0
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("⏰ 监听结束 (收到 %d 个事件)\n", eventCount)
			return
		case event, ok := <-eventChan:
			if !ok {
				fmt.Println("📞 事件通道关闭")
				return
			}
			eventCount++
			if event.IsValid() {
				fmt.Printf("📝 [事件 %d] 配置更新: %s:%d, DB: %s\n", eventCount,
					event.Config.Server.Host, event.Config.Server.Port, event.Config.Database.URL)
			} else {
				fmt.Printf("❌ [事件 %d] 配置更新失败: %v\n", eventCount, event.Error)
			}
		}
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// startWSServer starts WebSocket servers on both HTTP and HTTPS
func startWSServer() {
	// Start HTTP WebSocket server (for Load)
	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/config.xml", handleWebSocket)

		server := &http.Server{
			Addr:              ":8080",
			Handler:           mux,
			ReadHeaderTimeout: 5 * time.Second,
			ReadTimeout:       10 * time.Second,
			WriteTimeout:      10 * time.Second,
		}

		log.Printf("HTTP WebSocket server listening on :8080")
		if err := server.ListenAndServe(); err != nil {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	// Start HTTPS WebSocket server (for Subscribe)
	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/config.xml", handleSecureWebSocket)

		server := &http.Server{
			Addr:              ":8443",
			Handler:           mux,
			ReadHeaderTimeout: 5 * time.Second,
			ReadTimeout:       10 * time.Second,
			WriteTimeout:      10 * time.Second,
		}

		log.Printf("HTTPS WebSocket server listening on :8443")
		if err := server.ListenAndServeTLS("server.crt", "server.key"); err != nil {
			log.Printf("HTTPS server error: %v", err)
		}
	}()
}

// handleWebSocket handles regular WebSocket connections (for Load)
func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	// Send initial configuration and close (for Load operation)
	configXML := `<?xml version="1.0" encoding="UTF-8"?>
<config>
    <server>
        <host>localhost</host>
        <port>8080</port>
    </server>
    <database>
        <url>mysql://user:pass@localhost:3306/testdb</url>
    </database>
</config>`

	if err := conn.WriteMessage(websocket.TextMessage, []byte(configXML)); err != nil {
		log.Printf("Write error: %v", err)
	}
}

// handleSecureWebSocket handles secure WebSocket connections (for Subscribe)
func handleSecureWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Secure WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	log.Printf("Secure WebSocket client connected")

	// Send initial configuration
	configXML := `<?xml version="1.0" encoding="UTF-8"?>
<config>
    <server>
        <host>secure-host</host>
        <port>8443</port>
    </server>
    <database>
        <url>mysql://user:pass@secure-db:3306/testdb</url>
    </database>
</config>`

	if err := conn.WriteMessage(websocket.TextMessage, []byte(configXML)); err != nil {
		log.Printf("Initial write error: %v", err)
		return
	}

	// Send periodic updates
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	updateCount := 0
	for range ticker.C {
		updateCount++
		updatedXML := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<config>
    <server>
        <host>secure-host-v%d</host>
        <port>%d</port>
    </server>
    <database>
        <url>mysql://user:pass@secure-db-v%d:3306/testdb</url>
    </database>
</config>`, updateCount, 8443+updateCount, updateCount)

		if err := conn.WriteMessage(websocket.TextMessage, []byte(updatedXML)); err != nil {
			log.Printf("Update write error: %v", err)
			return
		}
		log.Printf("Sent config update #%d", updateCount)

		// Stop after 5 updates
		if updateCount >= 5 {
			return
		}
	}
}

// generateTLSCert generates a self-signed certificate for testing
func generateTLSCert() error {
	// Generate private key
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	// Create certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization:  []string{"Test Company"},
			Country:       []string{"US"},
			Province:      []string{""},
			Locality:      []string{"San Francisco"},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses: nil,
		DNSNames:    []string{"localhost"},
	}

	// Create certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return err
	}

	// Save certificate
	certOut, err := os.Create("server.crt")
	if err != nil {
		return err
	}
	defer certOut.Close()

	_ = pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	// Save private key
	keyOut, err := os.Create("server.key")
	if err != nil {
		return err
	}
	defer keyOut.Close()

	_ = pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})

	return nil
}
