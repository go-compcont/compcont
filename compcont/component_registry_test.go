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
	TypedCreateComponentFunc: func(registry IComponentContainer, config ConfigA) (component IComponentA, err error) {
		component = &ComponentA{
			ConfigA: config,
		}
		return
	},
}

type ConfigB struct {
	TestB  string                                     `ccf:"test_b"`
	ReferA TypedComponentConfig[ConfigA, IComponentA] `ccf:"refer_a"`
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
	TypedCreateComponentFunc: func(registry IComponentContainer, config ConfigB) (component IComponentB, err error) {
		componentA, err := config.ReferA.LoadComponent(registry)
		if err != nil {
			return
		}
		component = &ComponentB{
			ConfigB:    config,
			componentA: componentA,
		}
		return
	},
}

func Test(t *testing.T) {
	DefaultFactoryRegistry.Register(factoryA)
	DefaultFactoryRegistry.Register(factoryB)

	registry := NewComponentContainer()
	err := registry.LoadNamedComponents(map[ComponentName]ComponentConfig{
		"ca": TypedComponentConfig[ConfigA, IComponentA]{
			Type: "a",
			Config: ConfigA{
				TestA: "testa",
			},
		}.ToComponentConfig(),
		"cb": (&TypedComponentConfig[ConfigB, IComponentB]{
			Type: "b",
			Deps: []ComponentName{"ca"},
			Config: ConfigB{
				TestB:  "testb",
				ReferA: TypedComponentConfig[ConfigA, IComponentA]{Refer: "ca"},
			},
		}).ToComponentConfig(),
	})
	assert.NoError(t, err)

	err = registry.LoadNamedComponents(map[ComponentName]ComponentConfig{
		"cb1": {
			Type: "b",
			Deps: []ComponentName{"ca"},
			Config: map[string]any{
				"test_b":  "testb",
				"refer_a": map[string]any{"refer": "ca"},
			},
		},
	})
	assert.NoError(t, err)

	componentB, err := LoadComponent[IComponentB](registry, ComponentConfig{Refer: "cb"})
	assert.NoError(t, err)

	assert.Equal(t, "testb", componentB.GetConfigB().TestB)

	componentB1, err := LoadComponent[IComponentB](registry, ComponentConfig{Refer: "cb1"})
	assert.NoError(t, err)

	assert.Equal(t, "testb", componentB1.GetConfigB().TestB)
}
