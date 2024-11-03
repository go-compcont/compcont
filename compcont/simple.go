package compcont

import (
	"fmt"
	"reflect"
	"time"

	"github.com/mitchellh/mapstructure"
)

type TypedComponent[Instance any] struct {
	Context  Context
	Instance Instance
}

func GetComponent[Instance any](container IComponentContainer, name ComponentName) (ret TypedComponent[Instance], err error) {
	r, err := container.GetComponent(name)
	if err != nil {
		return
	}
	instance, ok := r.Instance.(Instance)
	if !ok {
		err = fmt.Errorf("get component failed, %w, name: %s, component type: %s, expected instance type %v, but got %v", ErrComponentTypeMismatch, name, r.Context.Config.Type, reflect.TypeOf(ret.Instance), reflect.TypeOf(r.Instance))
		return
	}
	ret = TypedComponent[Instance]{
		Context:  r.Context,
		Instance: instance,
	}
	return
}

// 根据指定类型加载一个组件实例
func LoadAnonymousComponent[Instance any](container IComponentContainer, config ComponentConfig) (ret TypedComponent[Instance], err error) {
	r, err := container.LoadAnonymousComponent(config)
	if err != nil {
		return
	}
	instance, ok := r.Instance.(Instance)
	if !ok {
		err = fmt.Errorf("%w, component type: %s, but expected %v", ErrComponentTypeMismatch, config.Type, reflect.TypeOf(ret))
		return
	}
	ret = TypedComponent[Instance]{
		Context:  r.Context,
		Instance: instance,
	}
	return
}

type CreateInstanceFunc func(ctx Context, config any) (instance any, err error)

type DestroyInstanceFunc func(ctx Context, instance any) (err error)

func decodeMapConfig[Config any](mapConfig map[string]any, structureConfig *Config) (err error) {
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName:     ConfigFieldTagName,
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

type TypedCreateInstanceFunc[Config any, Instance any] func(ctx Context, config Config) (instance Instance, err error)

func (f TypedCreateInstanceFunc[Config, Instance]) ToAny() CreateInstanceFunc {
	return func(ctx Context, rawConfig any) (comp any, err error) {
		switch v := rawConfig.(type) {
		case nil:
			var cfg Config
			return f(ctx, cfg)
		case Config:
			return f(ctx, v)
		case map[string]any:
			var cfg Config
			err = decodeMapConfig(v, &cfg)
			if err != nil {
				return
			}
			return f(ctx, cfg)
		default:
			err = fmt.Errorf("unexpected config type %s", reflect.ValueOf(rawConfig))
			return
		}
	}
}

type TypedDestroyInstanceFunc[Instance any] func(ctx Context, instance Instance) (err error)

func (f TypedDestroyInstanceFunc[Component]) ToAny() DestroyInstanceFunc {
	return func(ctx Context, component any) (err error) {
		if v, ok := component.(Component); ok {
			return f(ctx, v)
		}
		err = fmt.Errorf("unexpected component type %s", reflect.ValueOf(component))
		return
	}
}

type TypedComponentConfig[Config any, Component any] struct {
	Name   ComponentName   `json:"name" yaml:"name"`
	Type   ComponentTypeID `json:"type" yaml:"type"`     // 组件类型
	Refer  string          `json:"refer" yaml:"refer"`   // 来自其他组件的引用
	Deps   []ComponentName `json:"deps" yaml:"deps"`     // 构造该组件需要依赖的其他组件名称
	Config Config          `json:"config" yaml:"config"` // 组件的自身配置
}

func (c TypedComponentConfig[Config, Component]) ToAny() ComponentConfig {
	return ComponentConfig{
		Name:   c.Name,
		Type:   c.Type,
		Refer:  c.Refer,
		Deps:   c.Deps,
		Config: c.Config,
	}
}

func (c TypedComponentConfig[Config, Component]) LoadComponent(container IComponentContainer) (component TypedComponent[Component], err error) {
	return LoadAnonymousComponent[Component](container, c.ToAny())
}

type TypedSimpleComponentFactory[Config any, Component any] struct {
	TypeID              ComponentTypeID
	CreateInstanceFunc  TypedCreateInstanceFunc[Config, Component]
	DestroyInstanceFunc TypedDestroyInstanceFunc[Component]
}

func (s *TypedSimpleComponentFactory[Config, Component]) Type() ComponentTypeID {
	return s.TypeID
}

func (s *TypedSimpleComponentFactory[Config, Component]) CreateInstance(ctx Context, config any) (instance any, err error) {
	if s.CreateInstanceFunc == nil {
		return
	}
	return s.CreateInstanceFunc.ToAny()(ctx, config)
}

func (s *TypedSimpleComponentFactory[Config, Component]) DestroyInstance(ctx Context, instance any) (err error) {
	if s.DestroyInstanceFunc == nil {
		return
	}
	return s.DestroyInstanceFunc.ToAny()(ctx, instance)
}
