package compcont

type SimpleComponentFactory struct {
	TypeName             ComponentType
	CreateComponentFunc  CreateComponentFunc
	DestroyComponentFunc DestroyComponentFunc
}

func (s *SimpleComponentFactory) Type() ComponentType {
	return s.TypeName
}

func (s *SimpleComponentFactory) CreateComponent(registry IComponentRegistry, config any) (component any, err error) {
	if s.CreateComponentFunc == nil {
		return
	}
	return s.CreateComponentFunc(registry, config)
}

func (s *SimpleComponentFactory) DestroyComponent(registry IComponentRegistry, component any) (err error) {
	if s.DestroyComponentFunc == nil {
		return
	}
	return s.DestroyComponentFunc(registry, component)
}

type TypedSimpleComponentFactory[Config any, Component any] struct {
	TypeName                  ComponentType
	TypedCreateComponentFunc  TypedCreateComponentFunc[Config, Component]
	TypedDestroyComponentFunc TypedDestroyComponentFunc[Component]
}

func (s *TypedSimpleComponentFactory[Config, Component]) Type() ComponentType {
	return s.TypeName
}

func (s *TypedSimpleComponentFactory[Config, Component]) CreateComponent(registry IComponentRegistry, config any) (component any, err error) {
	if s.TypedCreateComponentFunc == nil {
		return
	}
	return TypedCreateComponent(s.TypedCreateComponentFunc)(registry, config)
}

func (s *TypedSimpleComponentFactory[Config, Component]) DestroyComponent(registry IComponentRegistry, component any) (err error) {
	if s.TypedDestroyComponentFunc == nil {
		return
	}
	return TypedDestoryComponent(s.TypedDestroyComponentFunc)(registry, component)
}
