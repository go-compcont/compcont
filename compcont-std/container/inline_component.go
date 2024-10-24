package container

import "github.com/go-compcont/compcont/compcont"

const InlineContainerType compcont.ComponentType = "std.inline-container"

type InlineContainerConfig struct {
	Components map[compcont.ComponentName]compcont.ComponentConfig `ccf:"components"`
}

func MustRegisterInlineContainer(r compcont.IFactoryRegistry) {
	r.Register(&compcont.TypedSimpleComponentFactory[InlineContainerConfig, compcont.IComponentContainer]{
		TypeName: InlineContainerType,
		CreateInstanceFunc: func(container compcont.IComponentContainer, config InlineContainerConfig) (instance compcont.IComponentContainer, err error) {
			instance = NewComponentContainer(WithFactoryRegistry(container.FactoryRegistry()))
			err = instance.LoadNamedComponents(config.Components)
			return
		},
	})
}

func init() {
	MustRegisterInlineContainer(compcont.DefaultFactoryRegistry)
}
