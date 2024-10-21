package compcont

import (
	"fmt"
	"reflect"
)

type ComponentType string

type ComponentConfig struct {
	Refer  string        `json:"refer" yaml:"refer"`   // 该组件引用的其他组件
	Type   ComponentType `json:"type" yaml:"type"`     // 组件类型
	Deps   []string      `json:"deps" yaml:"deps"`     // 构造该组件需要依赖的其他组件名称
	Config any           `json:"config" yaml:"config"` // 组件的自身配置
}

type TypedComponentConfig[CFG any, COMP any] struct {
	ComponentConfig
}

func (c *TypedComponentConfig[CFG, COMP]) GetComponent(cc IComponentContainer) (comp COMP, err error) {
	return GetComponentByConfig[COMP](cc, c.ComponentConfig)
}

type ComponentBuilder = func(cc IComponentContainer, cfg any) (comp any, err error) // 组件创建器
type ComponentCloser = func(cc IComponentContainer, comp any) error                 // 组件销毁器

type IComponentContainer interface {
	GetComponentByConfig(cfg ComponentConfig) (ret any, err error)          // 根据单个配置引用一个组件或实例化一个匿名组件
	LoadComponentsFromConfig(cfgMap map[string]ComponentConfig) (err error) // 根据配置实例化组件集合
	UnloadAllComponents() (err error)                                       // 卸载已存在的所有组件
	RegisterBuilder(componentType ComponentType, builder ComponentBuilder) IComponentContainer
	RegisterCloser(componentType ComponentType, closer ComponentCloser) IComponentContainer
	GetAllRegisteredComponentType() []ComponentType
}

func GetComponentByConfig[COMP any](container IComponentContainer, cfg ComponentConfig) (ret COMP, err error) {
	if cfg.Type == "" && cfg.Refer == "" {
		err = fmt.Errorf("%w, type and refer must be set, expected component type: %s", ErrComponentConfigInvalid, reflect.TypeOf(ret))
		return
	}
	r, err := container.GetComponentByConfig(cfg)
	if err != nil {
		return
	}
	ret, ok := r.(COMP)
	if !ok {
		err = fmt.Errorf("%w, component type: %s, but expected %v", ErrComponentTypeMismatch, cfg.Type, reflect.TypeOf(ret))
		return
	}
	return
}

func TypedRegisterBuilderAdapter[Config any, COMP any](typedBuilder func(cc IComponentContainer, cfg Config) (comp COMP, err error)) ComponentBuilder {
	return func(cc IComponentContainer, rawCfg any) (comp any, err error) {
		switch v := rawCfg.(type) {
		case nil:
			var cfg Config
			return typedBuilder(cc, cfg)
		case Config:
			return typedBuilder(cc, v)
		case map[string]any:
			var cfg Config
			err = decodeMapConfig(v, &cfg)
			if err != nil {
				return
			}
			return typedBuilder(cc, cfg)
		default:
			err = fmt.Errorf("unexpected config type %s", reflect.ValueOf(rawCfg))
			return
		}
	}
}
