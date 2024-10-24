package container

import "github.com/go-compcont/compcont/compcont"

const InlineContainerType compcont.ComponentType = "std.container-inline"

type ContainerInlineConfig struct {
	Components   map[compcont.ComponentName]compcont.ComponentConfig `ccf:"components"`
	ExportMapper map[compcont.ComponentName]compcont.ComponentName   `ccf:"export_mapper"`
}

func MustRegisterContainerInline(r compcont.IFactoryRegistry) {
	r.Register(&compcont.TypedSimpleComponentFactory[ContainerInlineConfig, compcont.IComponentContainer]{
		TypeName: InlineContainerType,
		CreateInstanceFunc: func(container compcont.IComponentContainer, config ContainerInlineConfig) (instance compcont.IComponentContainer, err error) {
			instance = NewComponentContainer(WithFactoryRegistry(container.FactoryRegistry()))
			err = instance.LoadNamedComponents(config.Components)

			for inParent, inChild := range config.ExportMapper {
				var comp compcont.Component
				comp, err = instance.GetComponent(inChild)
				if err != nil {
					return
				}
				err = instance.PutComponent(inParent, comp)
				if err != nil {
					return
				}
			}
			return
		},
	})
}

func init() {
	MustRegisterContainerInline(compcont.DefaultFactoryRegistry)
}
