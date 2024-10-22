package compcont

import "errors"

var (
	ErrComponentTypeNotRegistered  = errors.New("component type not registered")
	ErrComponentAlreadyExists      = errors.New("component already exists")
	ErrComponentNameNotFound       = errors.New("component name not found")
	ErrComponentTypeMismatch       = errors.New("component type mismatch")
	ErrComponentDependencyNotFound = errors.New("component dependency not found")
	ErrCircularDependency          = errors.New("circular dependency detected")
	ErrComponentConfigInvalid      = errors.New("component config invalid")
)
