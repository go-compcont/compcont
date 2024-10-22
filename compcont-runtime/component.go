package compcontrt

import (
	"errors"
	"fmt"
	"sync"

	"github.com/go-compcont/compcont/compcont"
)

type ComponentName = compcont.ComponentName
type ComponentType = compcont.ComponentType
type IFactoryRegistry = compcont.IFactoryRegistry
type IComponentContainer = compcont.IComponentContainer

type ComponentMetadata = compcont.ComponentMetadata
type ComponentConfig = compcont.ComponentConfig

var (
	ErrComponentNameNotFound       = errors.New("component name not found")
	ErrComponentDependencyNotFound = errors.New("component dependency not found")
)

type innerComponent struct {
	Instance     any                        // 组件的实例
	Dependencies map[ComponentName]struct{} // 组件的依赖关系
	Type         ComponentType
}

type ComponentRegistry struct {
	factoryRegistry IFactoryRegistry
	components      map[ComponentName]innerComponent
	mu              sync.RWMutex
}

// GetComponentMetadata implements IComponentContainer.
func (c *ComponentRegistry) GetComponentMetadata(name ComponentName) (meta ComponentMetadata, err error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	inner, ok := c.components[name]
	if !ok {
		err = ErrComponentNameNotFound
		return
	}
	meta = ComponentMetadata{
		Name:         name,
		Type:         inner.Type,
		Dependencies: inner.Dependencies,
	}
	return
}

// FactoryRegistry implements IComponentRegistry.
func (c *ComponentRegistry) FactoryRegistry() IFactoryRegistry {
	return c.factoryRegistry
}

// LoadAnonymousComponent 加载一个匿名组件，返回该组件实例，生命周期不由Registry控制，需要由该方法的调用方自行处理
func (c *ComponentRegistry) LoadAnonymousComponent(config ComponentConfig) (component any, err error) {
	// 直接引用
	if config.Refer != "" {
		var ok bool
		c.mu.RLock()
		innerComponent, ok := c.components[config.Refer]
		c.mu.RUnlock()

		if !ok {
			err = fmt.Errorf("%w: %s", ErrComponentNameNotFound, config.Refer)
			return
		}
		component = innerComponent.Instance
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
	factory, err := c.factoryRegistry.GetFactory(config.Type)
	if err != nil {
		return
	}

	// 构造组件
	component, err = factory.CreateComponent(c, config.Config)
	return
}

// LoadNamedComponents 加载一批具名组件，内部会自行根据拓扑排序顺序加载组件
func (c *ComponentRegistry) LoadNamedComponents(configMap map[ComponentName]ComponentConfig) (err error) {
	// 构建组件依赖图
	dagGraph := make(map[ComponentName]map[ComponentName]struct{})
	for name, cfg := range configMap {
		for _, dep := range cfg.Deps {
			if _, ok := dagGraph[name]; !ok {
				dagGraph[name] = make(map[ComponentName]struct{})
			}
			dagGraph[name][dep] = struct{}{}
		}
	}

	// 移除已存在的依赖关系
	for name, cfg := range configMap {
		var deps []ComponentName
		for _, dep := range cfg.Deps {
			c.mu.RLock()
			_, ok := c.components[dep]
			c.mu.RUnlock()
			if !ok {
				deps = append(deps, dep)
			}
		}
		cfg.Deps = deps
		configMap[name] = cfg
	}

	// 对新组件集合进行拓扑排序
	orders, err := topologicalSort(configMap)
	if err != nil {
		return
	}

	// 组件的顺序加载器，TODO 可以实现组件的并发启动优化
	for _, name := range orders {
		component, err := c.LoadAnonymousComponent(configMap[name])
		if err != nil {
			return err
		}
		c.mu.Lock()
		c.components[name] = innerComponent{
			Instance:     component,
			Dependencies: dagGraph[name],
			Type:         configMap[name].Type,
		}
		c.mu.Unlock()
	}
	return
}

// UnloadNamedComponents implements IComponentRegistry.
func (c *ComponentRegistry) UnloadNamedComponents(name []ComponentName, recursive bool) error {
	panic("unimplemented")
}

// LoadedComponentNames implements IComponentRegistry.
func (c *ComponentRegistry) LoadedComponentNames() (names []ComponentName) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for t := range c.components {
		names = append(names, t)
	}
	return
}

type options struct {
	factoryRegistry IFactoryRegistry
}

type optionsFunc func(o *options)

func WithFactoryRegistry(factoryRegistry IFactoryRegistry) optionsFunc {
	return func(o *options) {
		o.factoryRegistry = factoryRegistry
	}
}

func NewComponentContainer(optFns ...optionsFunc) (cr IComponentContainer) {
	var opt options
	for _, fn := range optFns {
		fn(&opt)
	}

	if opt.factoryRegistry == nil {
		opt.factoryRegistry = compcont.DefaultFactoryRegistry
	}
	return &ComponentRegistry{
		factoryRegistry: opt.factoryRegistry,
		components:      make(map[ComponentName]innerComponent),
	}
}
