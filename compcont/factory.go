package compcont

// 组件工厂的抽象
type IFactoryRegistry interface {
	Register(f IComponentFactory) error                            // 注册组件工厂
	Unregister(t ComponentTypeID) error                            // 取消注册组件工厂
	RegisteredComponentTypes() (types []ComponentTypeID)           // 获取所有已注册的组件工厂
	GetFactory(t ComponentTypeID) (f IComponentFactory, err error) // 根据组件类型获取组件工厂
}

func MustRegister(registry IFactoryRegistry, component IComponentFactory) {
	err := registry.Register(component)
	if err != nil {
		panic(err)
	}
}
