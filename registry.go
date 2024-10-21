package compcont

import (
	"fmt"
	"sync"
)

type ComponentRegistry struct {
	factories  map[ComponentType]IComponentFactory
	components map[ComponentName]any
	mu         sync.RWMutex
}

// LoadAnonymousComponent implements IComponentRegistry.
func (c *ComponentRegistry) LoadAnonymousComponent(config ComponentConfig) (component any, err error) {
	// 直接引用
	if config.Refer != "" {
		var ok bool
		c.mu.RLock()
		component, ok = c.components[config.Refer]
		c.mu.RUnlock()

		if !ok {
			err = fmt.Errorf("%w: %s", ErrComponentNameNotFound, config.Refer)
			return
		}
		return
	}

	// 检查依赖关系是否满足
	for _, dep := range config.Deps {
		if _, ok := c.components[dep]; !ok {
			err = fmt.Errorf("%w, dependency %s not found", ErrComponentDependencyNotFound, dep)
			return
		}
	}

	// 获取工厂
	c.mu.RLock()
	factory, ok := c.factories[config.Type]
	c.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("%w, component type: %s", ErrComponentTypeNotRegistered, config.Type)
	}

	// 构造组件
	c.mu.RLock()
	component, err = factory.CreateComponent(c, config.Config)
	c.mu.RUnlock()
	return
}

// LoadNamedComponents implements IComponentRegistry.
func (c *ComponentRegistry) LoadNamedComponents(configMap map[ComponentName]ComponentConfig) (err error) {
	orders, err := topologicalSort(configMap)
	if err != nil {
		return
	}
	for _, name := range orders {
		component, err := c.LoadAnonymousComponent(configMap[name])
		if err != nil {
			return err
		}
		c.mu.Lock()
		c.components[name] = component
		c.mu.Unlock()
	}
	return
}

// UnloadNamedComponents implements IComponentRegistry.
func (c *ComponentRegistry) UnloadNamedComponents(name []ComponentName, recursive bool) error {
	panic("unimplemented")
}

func NewComponentRegistry() IComponentRegistry {
	return &ComponentRegistry{
		factories:  make(map[ComponentType]IComponentFactory),
		components: make(map[ComponentName]any),
	}
}

// Register implements IComponentRegistry.
func (c *ComponentRegistry) Register(f IComponentFactory) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.factories[f.Type()] = f
}

// RegisteredComponentTypes implements IComponentRegistry.
func (c *ComponentRegistry) RegisteredComponentTypes() (types []ComponentType) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for t := range c.factories {
		types = append(types, t)
	}
	return
}
