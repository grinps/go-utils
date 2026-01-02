module github.com/grinps/go-utils/config/ext

go 1.21

require (
	github.com/go-viper/mapstructure/v2 v2.2.1
	github.com/grinps/go-utils/config v0.4.0
	github.com/grinps/go-utils/errext v0.8.0
)

require github.com/grinps/go-utils/telemetry v0.3.0 // indirect

replace github.com/grinps/go-utils/config => ../
