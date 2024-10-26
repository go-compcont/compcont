package compcont

import "errors"

var (
	ErrComponentAlreadyExists      = errors.New("component already exists")
	ErrComponentTypeMismatch       = errors.New("component type mismatch")
	ErrComponentConfigInvalid      = errors.New("component config invalid")
	ErrComponentNameNotFound       = errors.New("component name not found")
	ErrComponentNameInvalid        = errors.New("component name is invalid")
	ErrComponentDependencyNotFound = errors.New("component dependency not found")
	ErrComponentTypeNotRegistered  = errors.New("component type not registered")
	ErrCircularDependency          = errors.New("circular dependency detected")
)
