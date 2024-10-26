module github.com/go-compcont/compcont/compcont-contrib/compcont-redis

go 1.23.2

require (
	github.com/go-compcont/compcont/compcont v0.0.0
	github.com/redis/go-redis/v9 v9.7.0
)

require (
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
)

replace github.com/go-compcont/compcont/compcont v0.0.0 => ../../compcont
