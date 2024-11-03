package reloading

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/go-compcont/compcont/compcont"
	"gopkg.in/yaml.v3"
)

type ConfigType string

const (
	ConfigTypeAuto ConfigType = "auto" // 按照json->yaml的顺序依次尝试解析
	ConfigTypeYAML ConfigType = "yaml"
	ConfigTypeJSON ConfigType = "json"
)

type OnReloadingConfigListener[T any] interface {
	OnReloadingConfig(ctx context.Context, cfg T) error
}

type OnReloadingConfigListenerFunc[T any] func(ctx context.Context, cfg T) error

func (o OnReloadingConfigListenerFunc[T]) OnReloadingConfig(ctx context.Context, cfg T) error {
	return o(ctx, cfg)
}

type IReloadingConfig[T any] interface {
	LoadConfig(ctx context.Context) (T, error)
	AddOnReloadingConfigListener(listener OnReloadingConfigListener[T]) int
	RemoveOnReloadingConfigListener(id int)
	Close() error
}

type ReloadingConfigOption[T any] struct {
	Reloading    IReloading
	StaticConfig *T
	StructMode   bool
	ConfigType   ConfigType
}

func NewReloadingConfig[T any](opt ReloadingConfigOption[T]) IReloadingConfig[T] {
	if opt.ConfigType == "" {
		opt.ConfigType = ConfigTypeAuto
	}

	ret := &ReloadingConfig[T]{
		staticConfig: opt.StaticConfig,
		innerRaw:     opt.Reloading,
		configType:   opt.ConfigType,
		structMode:   opt.StructMode,
	}
	opt.Reloading.AddOnReloadingListener(OnReloadingListenerFunc(func(ctx context.Context, data []byte) error {
		ret.currentConfig = nil
		return nil
	}))
	return ret
}

type ReloadingConfig[T any] struct {
	staticConfig  *T
	currentConfig *T // 如果当前有值，则直接返回，否则获取时重新反序列化
	configType    ConfigType
	structMode    bool
	innerRaw      IReloading
}

func (r *ReloadingConfig[T]) LoadConfig(ctx context.Context) (cfg T, err error) {
	if r.staticConfig != nil {
		cfg = *r.staticConfig
		return
	}
	if r.currentConfig != nil {
		cfg = *r.currentConfig
		return
	}
	currentConfigVal, err := r.unmarshal(r.innerRaw.Load(ctx))
	if err != nil {
		return
	}
	cfg = currentConfigVal
	return
}

func (r *ReloadingConfig[T]) unmarshal(data []byte) (ret T, err error) {
	unmarshalYAML := func() (ret T, err error) {
		decoder := yaml.NewDecoder(bytes.NewReader(data))
		decoder.KnownFields(r.structMode)
		err = decoder.Decode(&ret)
		return
	}
	unmarshalJSON := func() (ret T, err error) {
		decoder := json.NewDecoder(bytes.NewReader(data))
		if r.structMode {
			decoder.DisallowUnknownFields()
		}
		err = decoder.Decode(&ret)
		return
	}
	switch r.configType {
	case ConfigTypeAuto:
		ret, err = unmarshalJSON()
		if err == nil {
			return
		}
		ret, err = unmarshalYAML()
		if err == nil {
			return
		}
		return
	case ConfigTypeYAML:
		return unmarshalYAML()
	case ConfigTypeJSON:
		return unmarshalJSON()
	default:
		panic("unreachable")
	}
}

func (r *ReloadingConfig[T]) AddOnReloadingConfigListener(listener OnReloadingConfigListener[T]) int {
	if r.innerRaw == nil {
		return -1
	}
	return r.innerRaw.AddOnReloadingListener(OnReloadingListenerFunc(func(ctx context.Context, data []byte) error {
		cfg, err := r.unmarshal(data)
		if err != nil {
			return err
		}
		return listener.OnReloadingConfig(ctx, cfg)
	}))
}

func (r *ReloadingConfig[T]) RemoveOnReloadingConfigListener(id int) {
	if r.innerRaw == nil {
		return
	}
	r.innerRaw.RemoveOnReloadingListener(id)
}

func (r *ReloadingConfig[T]) Close() error {
	return r.innerRaw.Close()
}

type ReloadingConfigConfig[T any] struct {
	StaticConfig *T                                              `ccf:"static_config"`
	StructMode   bool                                            `ccf:"struct_mode"`
	ConfigType   ConfigType                                      `ccf:"config_type"`
	Reloading    *compcont.TypedComponentConfig[any, IReloading] `ccf:"reloading"`
}

func (r *ReloadingConfigConfig[T]) Build(cc compcont.IComponentContainer) (rc IReloadingConfig[T], err error) {
	var reloading IReloading
	if r.Reloading != nil {
		reloadingComp, err1 := r.Reloading.LoadComponent(cc)
		if err1 != nil {
			err = err1
			return
		}
		reloading = reloadingComp.Instance
	}
	rc = NewReloadingConfig(ReloadingConfigOption[T]{
		Reloading:    reloading,
		StaticConfig: r.StaticConfig,
		StructMode:   r.StructMode,
		ConfigType:   r.ConfigType,
	})
	return
}
