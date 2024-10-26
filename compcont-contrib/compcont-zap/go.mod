module github.com/go-compcont/compcont/compcont-contrib/compcont-zap

go 1.23.2

replace github.com/go-compcont/compcont/compcont v0.0.0 => ../../compcont

require (
	github.com/go-compcont/compcont/compcont v0.0.0
	go.uber.org/zap v1.27.0
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
)

require (
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	go.uber.org/multierr v1.10.0 // indirect
)
