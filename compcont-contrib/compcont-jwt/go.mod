module github.com/go-compcont/compcont/compcont-contrib/compcont-jwt

go 1.23.2

replace github.com/go-compcont/compcont/compcont v0.0.0 => ../../compcont

require (
	github.com/brianvoe/sjwt v0.5.1
	github.com/go-compcont/compcont/compcont v0.0.0
)

require github.com/mitchellh/mapstructure v1.5.0 // indirect
