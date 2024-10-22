package compcont

import (
	"fmt"
	"reflect"
	"time"

	"github.com/mitchellh/mapstructure"
)

type TypedCreateComponentFunc[Config any, Component any] func(container IComponentContainer, config Config) (component Component, err error)

type TypedDestroyComponentFunc[Component any] func(container IComponentContainer, component Component) (err error)

// 将带完整类型的组件构造函数泛化成接口通用的构造函数
func TypedCreateComponent[Config any, Component any](typedConstructor TypedCreateComponentFunc[Config, Component]) CreateComponentFunc {
	return func(cc IComponentContainer, rawConfig any) (comp any, err error) {
		switch v := rawConfig.(type) {
		case nil:
			var cfg Config
			return typedConstructor(cc, cfg)
		case Config:
			return typedConstructor(cc, v)
		case map[string]any:
			var cfg Config
			err = decodeMapConfig(v, &cfg)
			if err != nil {
				return
			}
			return typedConstructor(cc, cfg)
		default:
			err = fmt.Errorf("unexpected config type %s", reflect.ValueOf(rawConfig))
			return
		}
	}
}

func decodeMapConfig[C any](mapConfig map[string]any, structureConfig *C) (err error) {
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName:     "ccf",
		ErrorUnused: true,            // 配置文件如果多余出未使用的字段，则报错
		ZeroFields:  true,            // decode前对传入的结构体清零
		Result:      structureConfig, // 目标结构体
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),     // 自动解析duration
			mapstructure.StringToTimeHookFunc(time.RFC3339), // 自动解析时间
		),
	})
	if err != nil {
		return
	}
	err = decoder.Decode(mapConfig)
	if err != nil {
		return
	}
	return
}

func TypedDestoryComponent[Component any](typedDestructor TypedDestroyComponentFunc[Component]) DestroyComponentFunc {
	return func(cc IComponentContainer, component any) (err error) {
		if v, ok := component.(Component); ok {
			return typedDestructor(cc, v)
		}
		err = fmt.Errorf("unexpected component type %s", reflect.ValueOf(component))
		return
	}
}

type TypedComponentConfig[Config any, Component any] struct {
	Refer  ComponentName   `json:"refer" yaml:"refer"`   // 该组件引用的其他组件
	Type   ComponentType   `json:"type" yaml:"type"`     // 组件类型
	Deps   []ComponentName `json:"deps" yaml:"deps"`     // 构造该组件需要依赖的其他组件名称
	Config Config          `json:"config" yaml:"config"` // 组件的自身配置
}

func (c TypedComponentConfig[Config, Component]) ToComponentConfig() ComponentConfig {
	return ComponentConfig{
		Refer:  c.Refer,
		Type:   c.Type,
		Deps:   c.Deps,
		Config: c.Config,
	}
}

func (c TypedComponentConfig[Config, Component]) LoadComponent(container IComponentContainer) (component Component, err error) {
	return LoadComponent[Component](container, c.ToComponentConfig())
}

func LoadComponent[Component any](container IComponentContainer, config ComponentConfig) (ret Component, err error) {
	if config.Type == "" && config.Refer == "" {
		err = fmt.Errorf("%w, type and refer must be set, expected component type: %s", ErrComponentConfigInvalid, reflect.TypeOf(ret))
		return
	}
	r, err := container.LoadAnonymousComponent(config)
	if err != nil {
		return
	}
	ret, ok := r.(Component)
	if !ok {
		err = fmt.Errorf("%w, component type: %s, but expected %v", ErrComponentTypeMismatch, config.Type, reflect.TypeOf(ret))
		return
	}
	return
}
