package compcont

import (
	"fmt"
	"maps"
	"slices"
	"sync"
)

type ComponentContainer struct {
	componentBuilders map[ComponentType]ComponentBuilder // 组件构造器中心
	componentClosers  map[ComponentType]ComponentCloser  // 组件销毁器中心
	components        map[string]any                     // 存放已构造好的全局容器
	componentTypes    map[string]ComponentType           // 组件名对应的组件类型
	mu                sync.RWMutex
}

func NewComponentContainer() *ComponentContainer {
	return &ComponentContainer{
		componentBuilders: make(map[ComponentType]ComponentBuilder),
		componentClosers:  make(map[ComponentType]ComponentCloser),
		components:        make(map[string]any),
		componentTypes:    make(map[string]ComponentType),
	}
}

// 注册组件的构造器
func (c *ComponentContainer) RegisterBuilder(componentType ComponentType, builder ComponentBuilder) IComponentContainer {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.componentBuilders[componentType] = builder
	return c
}

func (c *ComponentContainer) RegisterCloser(componentType ComponentType, closer ComponentCloser) IComponentContainer {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.componentClosers[componentType] = closer
	return c
}

func (c *ComponentContainer) UnloadAllComponents() (err error) {
	// TODO
	return
}

// 加载一个组件
func (c *ComponentContainer) loadComponentFromConfig(name string, cfg ComponentConfig) error {
	// 检查组件是否已经存在
	c.mu.RLock()
	_, ok := c.components[name]
	c.mu.RUnlock()
	if ok {
		return fmt.Errorf("%w, component name: %s", ErrComponentAlreadyExists, name)
	}

	// 检查组件依赖关系是否满足
	for _, dep := range cfg.Deps {
		c.mu.RLock()
		_, ok := c.components[dep]
		c.mu.RUnlock()
		if !ok {
			return fmt.Errorf("%w, dependency %s not found for component %s", ErrComponentDependencyNotFound, dep, name)
		}
	}

	// 获取组件构造器
	c.mu.RLock()
	cb, ok := c.componentBuilders[cfg.Type]
	c.mu.RUnlock()
	if !ok {
		return fmt.Errorf("%w, component type: %s", ErrComponentTypeNotRegistered, cfg.Type)
	}

	// 构造组件
	component, err := cb(c, cfg.Config)
	if err != nil {
		return fmt.Errorf("build component %s error: %w", name, err)
	}

	// 将组件添加到容器中
	c.mu.Lock()
	c.components[name] = component
	c.componentTypes[name] = cfg.Type
	c.mu.Unlock()
	return nil
}

// 注销一个组件
func (c *ComponentContainer) CloseComponent(name string) error {
	c.mu.RLock()
	component, ok := c.components[name]
	c.mu.RUnlock()
	if !ok {
		return fmt.Errorf("%w, component name: %s", ErrComponentNameNotFound, name)
	}

	// 获取组件销毁器
	c.mu.RLock()
	cp, ok := c.componentTypes[name]
	if !ok {
		panic(fmt.Errorf("component type not found for component name: %s", name))
	}
	closer, ok := c.componentClosers[cp]
	if !ok {
		panic(fmt.Errorf("component closer not found for component type: %s", cp))
	}
	c.mu.RUnlock()

	return closer(c, component)
}

func topologicalSort(cfgMap map[string]ComponentConfig) ([]string, error) {
	// 计算每个节点的入度
	inDegree := make(map[string]int)
	for name := range cfgMap {
		inDegree[name] = 0
	}
	for name, cfg := range cfgMap {
		for _, dep := range cfg.Deps {
			if _, ok := cfgMap[dep]; !ok {
				return nil, fmt.Errorf("component config error, %w, dependency %s not found for component %s", ErrComponentDependencyNotFound, dep, name)
			}
			inDegree[dep]++
		}
	}

	// 初始化队列，将所有入度为 0 的节点加入队列
	queue := []string{}
	for name, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, name)
		}
	}

	// 拓扑排序
	result := []string{}
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		result = append(result, node)

		for _, dep := range cfgMap[node].Deps {
			inDegree[dep]--
			if inDegree[dep] == 0 {
				queue = append(queue, dep)
			}
		}
	}

	// 检查是否有环
	if len(result) != len(cfgMap) {
		return nil, ErrCircularDependency
	}

	slices.Reverse(result)
	return result, nil
}

// 从配置map中构造并注入所有的组件
func (c *ComponentContainer) LoadComponentsFromConfig(cfgMap map[string]ComponentConfig) (err error) {
	// 按照依赖关系完成拓扑排序
	initOrders, err := topologicalSort(cfgMap)
	if err != nil {
		return
	}
	for _, name := range initOrders {
		if err := c.loadComponentFromConfig(name, cfgMap[name]); err != nil {
			return err
		}
	}
	return
}

// 构造一个新组件
func (c *ComponentContainer) newComponent(cfg ComponentConfig) (any, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	cb, ok := c.componentBuilders[cfg.Type]
	if !ok {
		return nil, fmt.Errorf("%w, component type: %s", ErrComponentTypeNotRegistered, cfg.Type)
	}

	return cb(c, cfg.Config)
}

// 从容器中获取组件
func (c *ComponentContainer) getComponent(name string) (ret any, err error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	component, ok := c.components[name]
	if !ok {
		err = fmt.Errorf("%w: %s", ErrComponentNameNotFound, name)
		return
	}

	ret = component
	return
}

// 尝试从容器中获取组件或者构造新组件
func (c *ComponentContainer) GetComponentByConfig(cfg ComponentConfig) (ret any, err error) {
	if cfg.Refer != "" {
		return c.getComponent(cfg.Refer)
	}
	return c.newComponent(cfg)
}

func (c *ComponentContainer) GetAllRegisteredComponentType() []ComponentType {
	s := slices.Collect(maps.Keys(c.componentBuilders))
	slices.Sort(s)
	return s
}
