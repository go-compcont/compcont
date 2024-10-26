package compcont

// 组件的容器抽象
type IComponentContainer interface {
	GetSelfComponentName() ComponentName                                            // 当容器自身作为组件时的组件名称，为空则为顶级组件
	FactoryRegistry() IFactoryRegistry                                              // 该组件容器所使用的组件工厂注册器
	LoadedComponentNames() (names []ComponentName)                                  // 获取所有已加载的组件名
	LoadNamedComponents(configMap map[ComponentName]ComponentConfig) error          // 实例化一批组件，内部自动基于拓扑排序的顺序完成组件的实例化
	UnloadNamedComponents(name []ComponentName, recursive bool) error               // 卸载一批组件，若指定recursive则递归地卸载依赖组件
	LoadAnonymousComponent(config ComponentConfig) (component Component, err error) // 立即加载一个匿名的组件
	GetComponent(name ComponentName) (component Component, err error)               // 获取一个已加载的具名组件
	PutComponent(name ComponentName, component Component) (err error)               // 直接放入一个组件
	GetParent() IComponentContainer                                                 // 如果是根容器，则返回nil
}
