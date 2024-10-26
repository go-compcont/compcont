package container

import (
	"testing"

	"github.com/go-compcont/compcont/compcont"
	"github.com/stretchr/testify/assert"
)

type ConfigA struct {
	TestA string `ccf:"test_a"`
}

type IComponentA interface {
	GetConfigA() ConfigA
}

type ComponentA struct {
	ConfigA
}

func (a *ComponentA) GetConfigA() ConfigA { return a.ConfigA }

var factoryA = &compcont.TypedSimpleComponentFactory[ConfigA, IComponentA]{
	TypeName: "a",
	CreateInstanceFunc: func(ctx compcont.Context, config ConfigA) (component IComponentA, err error) {
		component = &ComponentA{
			ConfigA: config,
		}
		return
	},
}

type ConfigB struct {
	TestB  string                                              `ccf:"test_b"`
	InnerA compcont.TypedComponentConfig[ConfigA, IComponentA] `ccf:"inner_a"`
}

type IComponentB interface {
	GetConfigB() ConfigB
}

type ComponentB struct {
	componentA IComponentA
	ConfigB
}

func (a *ComponentB) GetConfigB() ConfigB {
	return a.ConfigB
}

var factoryB = &compcont.TypedSimpleComponentFactory[ConfigB, IComponentB]{
	TypeName: "b",
	CreateInstanceFunc: func(ctx compcont.Context, config ConfigB) (component IComponentB, err error) {
		componentA, err := config.InnerA.LoadComponent(ctx.Container)
		if err != nil {
			return
		}
		component = &ComponentB{
			ConfigB:    config,
			componentA: componentA.Instance,
		}
		return
	},
}

func Test(t *testing.T) {
	compcont.DefaultFactoryRegistry.Register(factoryA)
	compcont.DefaultFactoryRegistry.Register(factoryB)

	registry := NewComponentContainer()
	err := registry.LoadNamedComponents(map[compcont.ComponentName]compcont.ComponentConfig{
		"cb": (&compcont.TypedComponentConfig[ConfigB, IComponentB]{
			Type: "b",
			Config: ConfigB{
				TestB: "testb",
				InnerA: compcont.TypedComponentConfig[ConfigA, IComponentA]{
					Type: "a",
					Config: ConfigA{
						TestA: "testa",
					},
				},
			},
		}).ToAny(),
	})
	assert.NoError(t, err)

	componentB, err := compcont.GetComponent[IComponentB](registry, "cb")
	assert.NoError(t, err)

	assert.Equal(t, "testa", componentB.Instance.GetConfigB().InnerA.Config.TestA)
}
