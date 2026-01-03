module github.com/grinps/go-utils/config/http

go 1.21

require (
	github.com/grinps/go-utils/config v0.5.0
	github.com/grinps/go-utils/errext v0.8.0
)

require github.com/grinps/go-utils/telemetry v0.3.0 // indirect

replace github.com/grinps/go-utils/config => ../
