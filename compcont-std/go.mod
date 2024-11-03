module github.com/go-compcont/compcont/compcont-std

go 1.23.1

require (
	github.com/go-compcont/compcont/compcont v0.0.0
	github.com/go-resty/resty/v2 v2.15.3
	github.com/stretchr/testify v1.9.0
	go.uber.org/zap v1.27.0
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	golang.org/x/net v0.27.0 // indirect
)

replace github.com/go-compcont/compcont/compcont v0.0.0 => ../compcont
