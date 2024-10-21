package compcont

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

type RedisConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
	Pass string `yaml:"pass"`
}

type Cache interface {
	Get(key string) (value string, err error)
}

type Redis struct {
	cfg RedisConfig
}

func (r *Redis) Get(key string) (value string, err error) {
	return
}

func redisBuilder(cc IComponentContainer, cfg any) (ret any, err error) {
	var redisCfg RedisConfig
	bs, err := yaml.Marshal(cfg)
	if err != nil {
		return
	}
	err = yaml.Unmarshal(bs, &redisCfg)
	if err != nil {
		return
	}
	ret = &Redis{
		cfg: redisCfg,
	}
	return
}

type RestyConfig struct {
	TimeoutMs int    `yaml:"timeout_ms"`
	BaseURL   string `yaml:"base_url"`
}

type Resty struct {
	cfg RestyConfig
}

func restyBuilder(cc IComponentContainer, cfg any) (ret any, err error) {
	var restyCfg RestyConfig
	bs, err := yaml.Marshal(cfg)
	if err != nil {
		return
	}
	err = yaml.Unmarshal(bs, &restyCfg)
	if err != nil {
		return
	}
	ret = &Resty{
		cfg: restyCfg,
	}
	return
}

const (
	ComponentTypeRedis = "redis"
	ComponentTypeResty = "resty"
)

func TestComponentContainer(t *testing.T) {
	cc := NewComponentContainer()
	cc.RegisterBuilder(ComponentTypeRedis, redisBuilder).
		RegisterBuilder(ComponentTypeResty, restyBuilder)
	err := cc.LoadComponentsFromConfig(map[string]ComponentConfig{
		"redis-0": {
			Type: "redis",
			Config: map[string]any{
				"host": "h1",
				"port": 2,
				"pass": "p",
			},
		},
		"resty-0": {
			Type: "resty",
			Deps: []string{"redis-0"},
			Config: map[string]any{
				"timeout_ms": 100,
			},
		},
	})
	assert.NoError(t, err)

	ret, err := cc.getComponent("redis-0")
	assert.NoError(t, err)
	assert.IsType(t, &Redis{}, ret)

	_, ok := ret.(Cache)
	assert.True(t, ok)

	ret, err = cc.getComponent("resty-0")
	assert.NoError(t, err)
	assert.IsType(t, &Resty{}, ret)

	ret, err = cc.getComponent("redis-1")
	assert.Error(t, err)
	assert.Nil(t, ret)

}

func TestTopologicalSort(t *testing.T) {
	// 正常情况
	t.Run("Normal Case", func(t *testing.T) {
		// 复杂依赖关系
		graph := map[string]ComponentConfig{
			"a1": {Deps: []string{"b1", "b2"}},
			"b1": {Deps: []string{"c1", "c2"}},
			"b2": {Deps: []string{"c2", "c3"}},
			"c1": {Deps: []string{"d1"}},
			"c2": {Deps: []string{"d1", "d2"}},
			"c3": {Deps: []string{"d2", "d3"}},
			"d1": {Deps: []string{}},
			"d2": {Deps: []string{}},
			"d3": {Deps: []string{}},
			"e1": {Deps: []string{"a1", "d1"}},
			"e2": {Deps: []string{"e1", "d2"}},
			"e3": {Deps: []string{"e2", "d3"}},
		}
		order, err := topologicalSort(graph)
		assert.NoError(t, err)

		orderSet := make(map[string]struct{})
		for _, name := range order {
			// 检查该节点是否能够加入
			deps := graph[name].Deps
			for _, dep := range deps {
				if _, ok := orderSet[dep]; !ok {
					t.Errorf("dependency %s not found", dep)
				}
			}
			orderSet[name] = struct{}{}
		}
	})

	// 依赖缺失情况
	t.Run("Dependency Not Found", func(t *testing.T) {
		_, err := topologicalSort(map[string]ComponentConfig{
			"a1": {
				Deps: []string{"b1", "b2"},
			},
			"b1": {
				Deps: []string{"c1"},
			},
			"b2": {
				Deps: []string{"c2"}, // c2 不存在
			},
		})
		assert.Error(t, err)
		assert.Equal(t, ErrComponentDependencyNotFound, errors.Unwrap(err))
	})

	// 循环依赖情况
	t.Run("Circular Dependency", func(t *testing.T) {
		_, err := topologicalSort(map[string]ComponentConfig{
			"a1": {
				Deps: []string{"b1"},
			},
			"b1": {
				Deps: []string{"a1"}, // 循环依赖
			},
		})
		assert.Error(t, err)
		assert.Equal(t, ErrCircularDependency, err)
	})
}
