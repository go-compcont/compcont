package compcont

// 组件工厂的抽象
type IFactoryRegistry interface {
	Register(f IComponentFactory)                                // 注册组件工厂
	RegisteredComponentTypes() (types []ComponentType)           // 获取所有已注册的组件工厂
	GetFactory(t ComponentType) (f IComponentFactory, err error) // 根据组件类型获取组件工厂
}
