module ws-xml-example

go 1.25.0

replace github.com/sower-proxy/conf => ../..

require (
	github.com/gorilla/websocket v1.5.4-0.20250319132907-e064f32e3674
	github.com/sower-proxy/conf v0.0.0-00010101000000-000000000000
)

require (
	github.com/go-viper/mapstructure/v2 v2.4.0 // indirect
	golang.org/x/net v0.38.0 // indirect
)
