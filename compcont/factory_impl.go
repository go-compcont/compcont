package compcont

import (
	"fmt"
	"sync"
)

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

var DefaultFactoryRegistry IFactoryRegistry = NewFactoryRegistry()