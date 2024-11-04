package container

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/go-compcont/compcont/compcont"
	"gopkg.in/yaml.v3"
)

const ContainerImportType compcont.ComponentTypeID = "std.container-import"

type ImportFileConfig map[string]compcont.ComponentConfig

type ContainerImportConfig struct {
	FromFile string `ccf:"from_file"` // 从外部文件导入配置
}

var importFactory compcont.IComponentFactory = &compcont.TypedSimpleComponentFactory[ContainerImportConfig, compcont.IComponentContainer]{
	TypeID: ContainerImportType,
	CreateInstanceFunc: func(ctx compcont.Context, config ContainerImportConfig) (instance compcont.IComponentContainer, err error) {
		instance = compcont.NewComponentContainer(
			compcont.WithFactoryRegistry(ctx.Container.FactoryRegistry()),
			compcont.WithParentContainer(ctx.Container),
			compcont.WithContext(ctx),
		)
		var bs []byte
		bs, err = os.ReadFile(config.FromFile)
		if err != nil {
			return
		}

		components := []compcont.ComponentConfig{}
		switch {
		case strings.HasSuffix(config.FromFile, ".json"):
			err = json.Unmarshal(bs, &components)
			if err != nil {
				return
			}
		case strings.HasSuffix(config.FromFile, ".yml") || strings.HasSuffix(config.FromFile, ".yaml"):
			err = yaml.Unmarshal(bs, &components)
			if err != nil {
				return
			}
		default:
			err = fmt.Errorf("unsupported config file format: %s", config.FromFile)
			return
		}
		err = instance.LoadNamedComponents(components)
		return
	},
}

func MustRegisterContainerImport(r compcont.IFactoryRegistry) {
	compcont.MustRegister(r, importFactory)
}

func init() {
	MustRegisterContainerImport(compcont.DefaultFactoryRegistry)
}
