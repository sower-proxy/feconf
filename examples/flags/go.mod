module flags-example

go 1.25.0

replace github.com/sower-proxy/feconf => ../..

require github.com/sower-proxy/feconf v0.0.0-00010101000000-000000000000

require (
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/go-viper/mapstructure/v2 v2.4.0 // indirect
	golang.org/x/sys v0.38.0 // indirect
)
