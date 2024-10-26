package compcont

import (
	"testing"

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

var factoryA = &TypedSimpleComponentFactory[ConfigA, IComponentA]{
	TypeName: "a",
	CreateInstanceFunc: func(ctx Context, config ConfigA) (component IComponentA, err error) {
		component = &ComponentA{
			ConfigA: config,
		}
		return
	},
}

type ConfigB struct {
	TestB  string                                     `ccf:"test_b"`
	InnerA TypedComponentConfig[ConfigA, IComponentA] `ccf:"inner_a"`
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

var factoryB = &TypedSimpleComponentFactory[ConfigB, IComponentB]{
	TypeName: "b",
	CreateInstanceFunc: func(ctx Context, config ConfigB) (component IComponentB, err error) {
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
	DefaultFactoryRegistry.Register(factoryA)
	DefaultFactoryRegistry.Register(factoryB)

	registry := NewComponentContainer()
	err := registry.LoadNamedComponents(map[ComponentName]ComponentConfig{
		"cb": (&TypedComponentConfig[ConfigB, IComponentB]{
			Type: "b",
			Config: ConfigB{
				TestB: "testb",
				InnerA: TypedComponentConfig[ConfigA, IComponentA]{
					Type: "a",
					Config: ConfigA{
						TestA: "testa",
					},
				},
			},
		}).ToAny(),
	})
	assert.NoError(t, err)

	componentB, err := GetComponent[IComponentB](registry, "cb")
	assert.NoError(t, err)

	assert.Equal(t, "testa", componentB.Instance.GetConfigB().InnerA.Config.TestA)
}
