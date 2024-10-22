package compcont

type ComponentType string

type ComponentName string

// 一个组件工厂
type IFactory interface {
	Type() ComponentType // 组件唯一类型名称
	// 组件创建器，这里并没有明确config应该到底是什么类型，可以放到具体实现上既可以是map也可以是struct
	CreateComponent(container IComponentContainer, config any) (component any, err error)
	DestroyComponent(container IComponentContainer, component any) (err error) // 组件销毁器
}

// 组件工厂注册器
type IFactoryRegistry interface {
	Register(f IFactory)                                // 注册组件工厂
	RegisteredComponentTypes() (types []ComponentType)  // 获取所有已注册的组件工厂
	GetFactory(t ComponentType) (f IFactory, err error) // 根据组件类型获取组件工厂
}

// 组件工厂的注册中心
type IComponentContainer interface {
	FactoryRegistry() IFactoryRegistry                                           // 该组件容器所使用的组件工厂注册器
	LoadedComponentNames() (names []ComponentName)                               // 获取所有已加载的组件名
	LoadNamedComponents(configMap map[ComponentName]ComponentConfig) error       // 实例化一批组件，内部自动基于拓扑排序的顺序完成组件的实例化
	UnloadNamedComponents(name []ComponentName, recursive bool) error            // 卸载一批组件，若指定recursive则递归地卸载依赖组件
	LoadAnonymousComponent(config ComponentConfig) (component any, err error)    // 立即加载一个匿名的component
	GetComponentMetadata(name ComponentName) (meta ComponentMetadata, err error) // 获取已加载组件的元数据
}

type CreateComponentFunc TypedCreateComponentFunc[any, any]

type DestroyComponentFunc TypedDestroyComponentFunc[any]

type ComponentConfig TypedComponentConfig[any, any]

type ComponentMetadata struct {
	Name         ComponentName
	Type         ComponentType
	Dependencies map[ComponentName]struct{}
}
