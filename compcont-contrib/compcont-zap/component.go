package compcontzap

import (
	"github.com/go-compcont/compcont/compcont"
	"go.uber.org/zap"
)

const TypeName compcont.ComponentType = "contrib.zap"

type LoggerProvider interface {
	GetLogger() *zap.Logger
}

type getLoggerFunc func() *zap.Logger

func (f getLoggerFunc) GetLogger() *zap.Logger {
	return f()
}

var factory compcont.IComponentFactory = &compcont.TypedSimpleComponentFactory[Config, LoggerProvider]{
	TypeName: TypeName,
	CreateInstanceFunc: func(ctx compcont.Context, config Config) (instance LoggerProvider, err error) {
		logger, err := New(config)
		if err != nil {
			return
		}
		instance = getLoggerFunc(func() *zap.Logger {
			return logger
		})
		return
	},
}

func MustRegister(registry compcont.IFactoryRegistry) {
	registry.Register(factory)
}

func init() {
	MustRegister(compcont.DefaultFactoryRegistry)
}
