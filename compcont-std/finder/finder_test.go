package finder

import (
	"fmt"
	"log/slog"
	"reflect"
	"testing"

	"github.com/go-compcont/compcont/compcont"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

var testComp compcont.IComponentFactory = &compcont.TypedSimpleComponentFactory[string, any]{
	TypeName: "echo",
	CreateInstanceFunc: func(ctx compcont.Context, config string) (instance any, err error) {
		instance = config
		slog.Info(
			"echo component",
			slog.String("absolute", fmt.Sprint(ctx.GetAbsolutePath())),
			slog.String("name", string(ctx.Name)),
			slog.String("container", reflect.TypeOf(ctx.Container).String()),
			slog.Any("instance", instance),
		)
		return
	},
}

var outputIns compcont.IComponentFactory = &compcont.TypedSimpleComponentFactory[compcont.ComponentConfig, any]{
	TypeName: "output",
	CreateInstanceFunc: func(ctx compcont.Context, config compcont.ComponentConfig) (instance any, err error) {
		s, err := compcont.LoadAnonymousComponent[any](ctx.Container, config)
		if err != nil {
			return
		}
		slog.Info(
			"output component",
			slog.String("absolute", fmt.Sprint(ctx.GetAbsolutePath())),
			slog.String("name", string(ctx.Name)),
			slog.String("container", reflect.TypeOf(ctx.Container).String()),
			slog.Any("output", s),
		)
		instance = s
		return
	},
}

const cfgYaml = `
test1:
  type: "echo"
  config: "Hello t1"

test2:
  type: "echo"
  config: "Hello t2"

test3:
  type: "echo"
  deps: [test1,test2]
  config: "Hello t3"

test4: { deps: [test1], refer: "test1" }

output_test4:
  type: "output"
  deps: ["test4"]
  config: { refer: "test4" }

c1:
  type: "std.container-inline"
  deps: [ "output_test4" ]
  config:
    components:
      test1:
        type: "echo"
        config: "Container t1"
      test2: { deps: [test1], refer: "test1" }
      output_test4:
        type: "output"
        deps: ["test2"]
        config: {refer: "test2"}

      finder_output:
        type: "output"
        config:
          type: "std.finder"
          config: "../output_test4"

c2:
  type: "std.container-import"
  deps: [c1]
  config:
    from_file: "test.yaml"
`

func TestFinder(t *testing.T) {
	cc := compcont.NewComponentContainer()
	compcont.DefaultFactoryRegistry.Register(testComp)
	compcont.DefaultFactoryRegistry.Register(outputIns)
	cfg := make(map[compcont.ComponentName]compcont.ComponentConfig)
	err := yaml.Unmarshal([]byte(cfgYaml), &cfg)
	assert.NoError(t, err)
	err = cc.LoadNamedComponents(cfg)
	assert.NoError(t, err)
}
