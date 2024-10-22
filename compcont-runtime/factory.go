package compcontrt

import (
	"errors"
	"fmt"
	"sync"

	"github.com/go-compcont/compcont/compcont"
)

type IFactory = compcont.IFactory

var ErrComponentTypeNotRegistered = errors.New("component type not registered")

type FactoryRegistry struct {
	factories map[ComponentType]IFactory
	mu        sync.RWMutex
}

func NewFactoryRegistry() IFactoryRegistry {
	return &FactoryRegistry{
		factories: make(map[ComponentType]IFactory),
	}
}

// Register implements IComponentFactoryRegistry.
func (c *FactoryRegistry) Register(f IFactory) {
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

func (c *FactoryRegistry) GetFactory(t ComponentType) (f IFactory, err error) {
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

func init() {
	compcont.DefaultFactoryRegistry = NewFactoryRegistry()
}
