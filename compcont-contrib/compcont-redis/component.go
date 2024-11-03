package compcontredis

import (
	"github.com/go-compcont/compcont/compcont"
	"github.com/redis/go-redis/v9"
)

const TypeID compcont.ComponentTypeID = "contrib.redis"

type Config struct {
	URL string `ccf:"url"`
}

type Component interface {
	GetClient() *redis.Client
	Destroy() error
}

type componentFunc struct {
	GetClientFunc func() *redis.Client
	DestroyFunc   func() error
}

func (f componentFunc) GetClient() *redis.Client {
	return f.GetClientFunc()
}
func (f componentFunc) Destroy() error {
	return f.DestroyFunc()
}

func New(cfg Config) (comp Component, err error) {
	options, err := redis.ParseURL(cfg.URL)
	if err != nil {
		return
	}
	rdb := redis.NewClient(options)
	comp = &componentFunc{
		GetClientFunc: func() *redis.Client { return rdb },
		DestroyFunc:   rdb.Close,
	}
	return
}

var factory compcont.IComponentFactory = &compcont.TypedSimpleComponentFactory[Config, Component]{
	TypeID: TypeID,
	CreateInstanceFunc: func(ctx compcont.Context, config Config) (instance Component, err error) {
		return New(config)
	},
	DestroyInstanceFunc: func(ctx compcont.Context, instance Component) (err error) {
		return instance.Destroy()
	},
}

func MustRegister(registry compcont.IFactoryRegistry) {
	registry.Register(factory)
}

func init() {
	MustRegister(compcont.DefaultFactoryRegistry)
}
