module redis-ini-example

go 1.26.0

replace github.com/sower-proxy/feconf => ../..

require github.com/sower-proxy/feconf v0.0.0-00010101000000-000000000000

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/go-redis/redis/v8 v8.11.5 // indirect
	github.com/go-viper/mapstructure/v2 v2.5.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	gopkg.in/ini.v1 v1.67.3 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
