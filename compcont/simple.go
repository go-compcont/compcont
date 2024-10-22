package compcont

type SimpleComponentFactory struct {
	TypeName             ComponentType
	CreateComponentFunc  CreateComponentFunc
	DestroyComponentFunc DestroyComponentFunc
}

func (s *SimpleComponentFactory) Type() ComponentType {
	return s.TypeName
}

func (s *SimpleComponentFactory) CreateComponent(container IComponentContainer, config any) (component any, err error) {
	if s.CreateComponentFunc == nil {
		return
	}
	return s.CreateComponentFunc(container, config)
}

func (s *SimpleComponentFactory) DestroyComponent(container IComponentContainer, component any) (err error) {
	if s.DestroyComponentFunc == nil {
		return
	}
	return s.DestroyComponentFunc(container, component)
}

type TypedSimpleComponentFactory[Config any, Component any] struct {
	TypeName                  ComponentType
	TypedCreateComponentFunc  TypedCreateComponentFunc[Config, Component]
	TypedDestroyComponentFunc TypedDestroyComponentFunc[Component]
}

func (s *TypedSimpleComponentFactory[Config, Component]) Type() ComponentType {
	return s.TypeName
}

func (s *TypedSimpleComponentFactory[Config, Component]) CreateComponent(container IComponentContainer, config any) (component any, err error) {
	if s.TypedCreateComponentFunc == nil {
		return
	}
	return TypedCreateComponent(s.TypedCreateComponentFunc)(container, config)
}

func (s *TypedSimpleComponentFactory[Config, Component]) DestroyComponent(container IComponentContainer, component any) (err error) {
	if s.TypedDestroyComponentFunc == nil {
		return
	}
	return TypedDestoryComponent(s.TypedDestroyComponentFunc)(container, component)
}
