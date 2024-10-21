package compcont

// 全局默认的组件工厂注册中心
var DefaultComponentRegistry IComponentRegistry = nil

type ComponentType string

type ComponentName string

// 一个组件工厂
type IComponentFactory interface {
	Type() ComponentType // 组件唯一类型名称
	// 组件创建器，这里并没有明确config应该到底是什么类型，可以放到具体实现上既可以是map也可以是struct
	CreateComponent(registry IComponentRegistry, config any) (component any, err error)
	DestroyComponent(registry IComponentRegistry, component any) (err error) // 组件销毁器
}

// 组件工厂的注册中心
type IComponentRegistry interface {
	Register(c IComponentFactory)                                             // 注册组件
	RegisteredComponentTypes() (types []ComponentType)                        // 获取所有已注册的组件
	LoadNamedComponents(configMap map[ComponentName]ComponentConfig) error    // 实例化一批组件，内部自动基于拓扑排序的顺序完成组件的实例化
	UnloadNamedComponents(name []ComponentName, recursive bool) error         // 卸载一批组件，若指定recursive则递归地卸载依赖组件
	LoadAnonymousComponent(config ComponentConfig) (componcnt any, err error) // 立即加载一个匿名的component
}

type CreateComponentFunc TypedCreateComponentFunc[any, any]

type DestroyComponentFunc TypedDestroyComponentFunc[any]

type ComponentConfig TypedComponentConfig[any, any]
