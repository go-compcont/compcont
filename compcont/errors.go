package compcont

import "errors"

var (
	ErrComponentAlreadyExists = errors.New("component already exists")
	ErrComponentTypeMismatch  = errors.New("component type mismatch")
	ErrComponentConfigInvalid = errors.New("component config invalid")
)
