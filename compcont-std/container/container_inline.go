package container

import "github.com/go-compcont/compcont/compcont"

const InlineContainerType compcont.ComponentType = "std.container-inline"

type ContainerInlineConfig struct {
	Components map[compcont.ComponentName]compcont.ComponentConfig `ccf:"components"`
}

func MustRegisterContainerInline(r compcont.IFactoryRegistry) {
	r.Register(&compcont.TypedSimpleComponentFactory[ContainerInlineConfig, compcont.IComponentContainer]{
		TypeName: InlineContainerType,
		CreateInstanceFunc: func(ctx compcont.Context, config ContainerInlineConfig) (instance compcont.IComponentContainer, err error) {
			instance = compcont.NewComponentContainer(
				compcont.WithParentContainer(ctx.Container),
				compcont.WithFactoryRegistry(ctx.Container.FactoryRegistry()),
				compcont.WithContext(ctx),
			)
			err = instance.LoadNamedComponents(config.Components)
			return
		},
	})
}

func init() {
	MustRegisterContainerInline(compcont.DefaultFactoryRegistry)
}
