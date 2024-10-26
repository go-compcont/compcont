package recovery

import (
	"github.com/gin-gonic/gin"
	"github.com/go-compcont/compcont/compcont"
)

const TypeName compcont.ComponentType = "contrib.gin-middleware-recovery"

type Config struct{}

var factory compcont.IComponentFactory = &compcont.TypedSimpleComponentFactory[Config, gin.HandlerFunc]{
	TypeName: TypeName,
	CreateInstanceFunc: func(ctx compcont.Context, config Config) (instance gin.HandlerFunc, err error) {
		instance = gin.Recovery()
		return
	},
}

func MustRegister(registry compcont.IFactoryRegistry) {
	registry.Register(factory)
}

func init() {
	MustRegister(compcont.DefaultFactoryRegistry)
}
