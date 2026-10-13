module flags-example

go 1.26.0

replace github.com/sower-proxy/feconf => ../..

require github.com/sower-proxy/feconf v0.0.0-00010101000000-000000000000

require (
	github.com/fsnotify/fsnotify v1.10.1 // indirect
	github.com/go-viper/mapstructure/v2 v2.5.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	golang.org/x/sys v0.46.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
