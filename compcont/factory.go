package compcont

import (
	"fmt"
	"sync"
)

var DefaultFactoryRegistry IFactoryRegistry = NewFactoryRegistry()

// 组件工厂的抽象
type IFactoryRegistry interface {
	Register(f IComponentFactory)                                // 注册组件工厂
	RegisteredComponentTypes() (types []ComponentType)           // 获取所有已注册的组件工厂
	GetFactory(t ComponentType) (f IComponentFactory, err error) // 根据组件类型获取组件工厂
}

type FactoryRegistry struct {
	factories map[ComponentType]IComponentFactory
	mu        sync.RWMutex
}

func NewFactoryRegistry() IFactoryRegistry {
	return &FactoryRegistry{
		factories: make(map[ComponentType]IComponentFactory),
	}
}

// Register implements IComponentFactoryRegistry.
func (c *FactoryRegistry) Register(f IComponentFactory) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.factories[f.Type()] = f
}

// RegisteredComponentTypes implements IComponentFactoryRegistry.
func (c *FactoryRegistry) RegisteredComponentTypes() (types []ComponentType) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for t := range c.factories {
		types = append(types, t)
	}
	return
}

func (c *FactoryRegistry) GetFactory(t ComponentType) (f IComponentFactory, err error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var ok bool
	f, ok = c.factories[t]
	if !ok {
		err = fmt.Errorf("%w, component type: %s", ErrComponentTypeNotRegistered, t)
		return
	}
	return
}
