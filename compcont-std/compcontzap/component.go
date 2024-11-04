package compcontzap

import (
	"github.com/go-compcont/compcont/compcont"
	"go.uber.org/zap"
)

const TypeID compcont.ComponentTypeID = "contrib.zap"

var factory compcont.IComponentFactory = &compcont.TypedSimpleComponentFactory[Config, *zap.Logger]{
	TypeID: TypeID,
	CreateInstanceFunc: func(ctx compcont.Context, config Config) (instance *zap.Logger, err error) {
		logger, err := New(config)
		if err != nil {
			return
		}
		instance = logger
		return
	},
}

func MustRegister(registry compcont.IFactoryRegistry) {
	compcont.MustRegister(registry, factory)
}

func init() {
	MustRegister(compcont.DefaultFactoryRegistry)
}
