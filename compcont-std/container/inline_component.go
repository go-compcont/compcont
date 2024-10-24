package container

import "github.com/go-compcont/compcont/compcont"

const InlineContainerType compcont.ComponentType = "std.container-inline"

type ContainerInlineConfig struct {
	Components map[compcont.ComponentName]compcont.ComponentConfig `ccf:"components"`
}

func MustRegisterContainerInline(r compcont.IFactoryRegistry) {
	r.Register(&compcont.TypedSimpleComponentFactory[ContainerInlineConfig, compcont.IComponentContainer]{
		TypeName: InlineContainerType,
		CreateInstanceFunc: func(container compcont.IComponentContainer, config ContainerInlineConfig) (instance compcont.IComponentContainer, err error) {
			instance = NewComponentContainer(WithFactoryRegistry(container.FactoryRegistry()))
			err = instance.LoadNamedComponents(config.Components)
			return
		},
	})
}

func init() {
	MustRegisterContainerInline(compcont.DefaultFactoryRegistry)
}
