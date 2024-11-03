package pprof

import (
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/go-compcont/compcont/compcont"
)

type Config struct {
	Gin         compcont.TypedComponentConfig[any, gin.IRouter] `ccf:"gin"`
	RoutePrefix string                                          `ccf:"route_prefix"`
}

const TypeID compcont.ComponentTypeID = "contrib.gin-pprof"

func New(cc compcont.IComponentContainer, cfg Config) (err error) {
	g, err := cfg.Gin.LoadComponent(cc)
	if err != nil {
		return
	}
	var options []string
	if len(cfg.RoutePrefix) > 0 {
		options = append(options, cfg.RoutePrefix)
	}
	pprof.RouteRegister(g.Instance, options...)
	return
}

var factory compcont.IComponentFactory = &compcont.TypedSimpleComponentFactory[Config, any]{
	TypeID: TypeID,
	CreateInstanceFunc: func(ctx compcont.Context, config Config) (instance any, err error) {
		err = New(ctx.Container, config)
		return
	},
}

func MustRegister(registry compcont.IFactoryRegistry) {
	registry.Register(factory)
}

func init() {
	MustRegister(compcont.DefaultFactoryRegistry)
}
