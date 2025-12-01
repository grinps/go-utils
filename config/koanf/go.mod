module github.com/grinps/go-utils/config/koanf

go 1.23.0

require (
	github.com/grinps/go-utils/config v0.0.0
	github.com/grinps/go-utils/errext v0.8.0
	github.com/knadh/koanf/v2 v2.1.1
)

require (
	github.com/go-viper/mapstructure/v2 v2.0.0-alpha.1 // indirect
	github.com/knadh/koanf/maps v0.1.2 // indirect
	github.com/knadh/koanf/parsers/json v1.0.0 // indirect
	github.com/knadh/koanf/providers/confmap v1.0.0 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
)

replace github.com/grinps/go-utils/config => ../
