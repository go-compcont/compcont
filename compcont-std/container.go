package compcontstd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/go-compcont/compcont/compcont"
	compcontrt "github.com/go-compcont/compcont/compcont-runtime"
	"gopkg.in/yaml.v3"
)

type ContainerConfig struct {
	FromFile   string                                              `ccf:"from_file"`
	Components map[compcont.ComponentName]compcont.ComponentConfig `ccf:"components"`
}

func (c *ContainerConfig) GetComponents() (components map[compcont.ComponentName]compcont.ComponentConfig, err error) {
	if len(c.Components) > 0 {
		components = c.Components
		return
	}
	if c.FromFile != "" {
		var bs []byte
		bs, err = os.ReadFile(c.FromFile)
		if err != nil {
			return
		}

		components = make(map[compcont.ComponentName]compcont.ComponentConfig)
		switch {
		case strings.HasSuffix(c.FromFile, ".json"):
			err = json.Unmarshal(bs, &components)
			if err != nil {
				return
			}
		case strings.HasSuffix(c.FromFile, ".yml") || strings.HasSuffix(c.FromFile, ".yaml"):
			err = yaml.Unmarshal(bs, &components)
			if err != nil {
				return
			}
		default:
			err = fmt.Errorf("unsupported config file format: %s", c.FromFile)
			return
		}
		return
	}
	return
}

type Container struct {
	parent compcont.IComponentContainer
	compcont.IComponentContainer
}

var containerFactory = &compcont.TypedSimpleComponentFactory[ContainerConfig, compcont.IComponentContainer]{
	TypeName: "std.container",
	TypedCreateComponentFunc: func(container compcont.IComponentContainer, config ContainerConfig) (component compcont.IComponentContainer, err error) {
		componentConfigs, err := config.GetComponents()
		if err != nil {
			return
		}
		component = &Container{
			IComponentContainer: compcontrt.NewComponentContainer(compcontrt.WithFactoryRegistry(container.FactoryRegistry())),
			parent:              container,
		}
		err = component.LoadNamedComponents(componentConfigs)
		return
	},
	TypedDestroyComponentFunc: func(registry, component compcont.IComponentContainer) (err error) {
		return component.UnloadNamedComponents(component.LoadedComponentNames(), true)
	},
}

type ReferParentConfig struct {
	Refer compcont.ComponentName `ccf:"refer"`
}

var referParentFactory = &compcont.TypedSimpleComponentFactory[ReferParentConfig, any]{
	TypeName: "std.refer-parent",
	TypedCreateComponentFunc: func(registry compcont.IComponentContainer, config ReferParentConfig) (component any, err error) {
		v, ok := registry.(*Container)
		if !ok {
			err = errors.New("std.refer-parent only can be used in std.container")
			return
		}
		return v.parent.LoadAnonymousComponent(compcont.ComponentConfig{Refer: config.Refer})
	},
}

type ReferChildConfig struct {
	Container compcont.ComponentConfig `ccf:"container"`
	Refer     compcont.ComponentName   `ccf:"refer"`
}

var referChildFactory = &compcont.TypedSimpleComponentFactory[ReferChildConfig, any]{
	TypeName: "std.refer-child",
	TypedCreateComponentFunc: func(registry compcont.IComponentContainer, config ReferChildConfig) (component any, err error) {
		child, err := compcont.LoadComponent[compcont.IComponentContainer](registry, config.Container)
		if err != nil {
			return
		}
		return child.LoadAnonymousComponent(compcont.ComponentConfig{Refer: config.Refer})
	},
}

func MustRegister(r compcont.IFactoryRegistry) {
	r.Register(containerFactory)
	r.Register(referParentFactory)
	r.Register(referChildFactory)
}

func init() {
	MustRegister(compcont.DefaultFactoryRegistry)
}
