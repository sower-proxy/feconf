module flags-example

go 1.25.0

replace github.com/sower-proxy/conf => ../..

require github.com/sower-proxy/conf v0.0.0-00010101000000-000000000000

require (
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/go-redis/redis/v8 v8.11.5 // indirect
	github.com/go-viper/mapstructure/v2 v2.4.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
)
