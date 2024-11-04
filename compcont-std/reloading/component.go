package reloading

import (
	"github.com/go-compcont/compcont/compcont"
	"github.com/go-resty/resty/v2"
)

const TypeID compcont.ComponentTypeID = "std.reloading"

var factory compcont.IComponentFactory = &compcont.TypedSimpleComponentFactory[Config, IReloading]{
	TypeID: TypeID,
	CreateInstanceFunc: func(ctx compcont.Context, cfg Config) (instance IReloading, err error) {
		var restyClient *resty.Client
		if cfg.Resty != nil {
			restyClientComp, err1 := cfg.Resty.LoadComponent(ctx.Container)
			if err1 != nil {
				err = err1
				return
			}
			restyClient = restyClientComp.Instance
		}
		instance = NewReloading(cfg, restyClient)
		return
	},
}

func MustRegister(registry compcont.IFactoryRegistry) {
	compcont.MustRegister(registry, factory)
}

func init() {
	MustRegister(compcont.DefaultFactoryRegistry)
}
