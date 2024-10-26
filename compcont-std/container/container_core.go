package container

import (
	"fmt"
	"sync"

	"github.com/go-compcont/compcont/compcont"
)

type ComponentContainer struct {
	selfName        compcont.ComponentName
	parent          compcont.IComponentContainer
	factoryRegistry compcont.IFactoryRegistry
	components      map[compcont.ComponentName]compcont.Component
	mu              sync.RWMutex
}

// GetSelfComponentName implements compcont.IComponentContainer.
func (c *ComponentContainer) GetSelfComponentName() compcont.ComponentName {
	return c.selfName
}

// GetParent implements compcont.IComponentContainer.
func (c *ComponentContainer) GetParent() compcont.IComponentContainer {
	return c.parent
}

// GetComponentMetadata implements IComponentContainer.
func (c *ComponentContainer) GetComponent(name compcont.ComponentName) (component compcont.Component, err error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	inner, ok := c.components[name]
	if !ok {
		err = fmt.Errorf("%w, name: %s", compcont.ErrComponentNameNotFound, name)
		return
	}
	component = inner
	return
}

// FactoryRegistry implements IComponentRegistry.
func (c *ComponentContainer) FactoryRegistry() compcont.IFactoryRegistry {
	return c.factoryRegistry
}

func (c *ComponentContainer) loadComponent(name compcont.ComponentName, config compcont.ComponentConfig) (component compcont.Component, err error) {
	if config.Type == "" {
		if config.Refer != "" { // 引用组件
			return c.GetComponent(config.Refer)
		}
		err = fmt.Errorf("%w, type && refer are empty", compcont.ErrComponentConfigInvalid)
		return
	}
	// 检查依赖关系是否满足
	for _, dep := range config.Deps {
		if _, ok := c.components[dep]; !ok {
			err = fmt.Errorf("%w, dependency %s not found", compcont.ErrComponentDependencyNotFound, dep)
			return
		}
	}

	// 获取工厂
	factory, err := c.factoryRegistry.GetFactory(config.Type)
	if err != nil {
		return
	}

	// 构造组件实例
	instance, err := factory.CreateInstance(compcont.Context{
		Name:      name,
		Container: c,
	}, config.Config)
	if err != nil {
		return
	}

	// 构造依赖
	deps := make(map[compcont.ComponentName]struct{})
	for _, dep := range config.Deps {
		deps[dep] = struct{}{}
	}

	// 构造组件
	component = compcont.Component{
		Type:         factory.Type(),
		Dependencies: deps,
		Instance:     instance,
	}
	return
}

// LoadAnonymousComponent 加载一个匿名组件，返回该组件实例，生命周期不由Registry控制，需要由该方法的调用方自行处理
func (c *ComponentContainer) LoadAnonymousComponent(config compcont.ComponentConfig) (component compcont.Component, err error) {
	return c.loadComponent("", config)
}

// PutComponent implements compcont.IComponentContainer.
func (c *ComponentContainer) PutComponent(name compcont.ComponentName, component compcont.Component) (err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.components[name] = component
	return
}

// LoadNamedComponents 加载一批具名组件，内部会自行根据拓扑排序顺序加载组件
func (c *ComponentContainer) LoadNamedComponents(configMap map[compcont.ComponentName]compcont.ComponentConfig) (err error) {
	// 校验组件名称
	for name := range configMap {
		if !name.Validate() {
			return fmt.Errorf("%w, name: %s", compcont.ErrComponentNameInvalid, name)
		}
	}
	// 构建组件依赖图
	dagGraph := make(map[compcont.ComponentName]map[compcont.ComponentName]struct{})
	for name, cfg := range configMap {
		for _, dep := range cfg.Deps {
			if _, ok := dagGraph[name]; !ok {
				dagGraph[name] = make(map[compcont.ComponentName]struct{})
			}
			dagGraph[name][dep] = struct{}{}
		}
	}

	// 移除已存在的依赖关系
	for name, cfg := range configMap {
		var deps []compcont.ComponentName
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
		component, err := c.loadComponent(name, configMap[name])
		if err != nil {
			return err
		}
		c.mu.Lock()
		c.components[name] = compcont.Component{
			Instance:     component.Instance,
			Dependencies: dagGraph[name],
			Type:         configMap[name].Type,
		}
		c.mu.Unlock()
	}
	return
}

// UnloadNamedComponents implements IComponentRegistry.
func (c *ComponentContainer) UnloadNamedComponents(name []compcont.ComponentName, recursive bool) error {
	panic("unimplemented")
}

// LoadedComponentNames implements IComponentRegistry.
func (c *ComponentContainer) LoadedComponentNames() (names []compcont.ComponentName) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for t := range c.components {
		names = append(names, t)
	}
	return
}

type options struct {
	factoryRegistry compcont.IFactoryRegistry
	parent          compcont.IComponentContainer
	selfName        compcont.ComponentName
}

type optionsFunc func(o *options)

func WithFactoryRegistry(factoryRegistry compcont.IFactoryRegistry) optionsFunc {
	return func(o *options) {
		o.factoryRegistry = factoryRegistry
	}
}

func WithParentContainer(parent compcont.IComponentContainer) optionsFunc {
	return func(o *options) {
		o.parent = parent
	}
}

func WithSelfNodeName(selfName compcont.ComponentName) optionsFunc {
	return func(o *options) {
		o.selfName = selfName
	}
}

func NewComponentContainer(optFns ...optionsFunc) (cr compcont.IComponentContainer) {
	var opt options
	for _, fn := range optFns {
		fn(&opt)
	}

	if opt.factoryRegistry == nil {
		opt.factoryRegistry = compcont.DefaultFactoryRegistry
	}
	return &ComponentContainer{
		selfName:        opt.selfName,
		factoryRegistry: opt.factoryRegistry,
		parent:          opt.parent,
		components:      make(map[compcont.ComponentName]compcont.Component),
	}
}
