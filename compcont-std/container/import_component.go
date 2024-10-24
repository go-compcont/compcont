package container

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/go-compcont/compcont/compcont"
	"gopkg.in/yaml.v3"
)

const ContainerImportType compcont.ComponentType = "std.container-import"

type ImportFileConfig map[string]compcont.ComponentConfig

type ContainerImportConfig struct {
	Import string `ccf:"import"`
}

func MustRegisterContainerImport(r compcont.IFactoryRegistry) {
	r.Register(&compcont.TypedSimpleComponentFactory[ContainerImportConfig, compcont.IComponentContainer]{
		TypeName: ContainerImportType,
		CreateInstanceFunc: func(container compcont.IComponentContainer, config ContainerImportConfig) (instance compcont.IComponentContainer, err error) {
			instance = NewComponentContainer(WithFactoryRegistry(container.FactoryRegistry()))
			var bs []byte
			bs, err = os.ReadFile(config.Import)
			if err != nil {
				return
			}

			components := make(map[compcont.ComponentName]compcont.ComponentConfig)
			switch {
			case strings.HasSuffix(config.Import, ".json"):
				err = json.Unmarshal(bs, &components)
				if err != nil {
					return
				}
			case strings.HasSuffix(config.Import, ".yml") || strings.HasSuffix(config.Import, ".yaml"):
				err = yaml.Unmarshal(bs, &components)
				if err != nil {
					return
				}
			default:
				err = fmt.Errorf("unsupported config file format: %s", config.Import)
				return
			}
			err = instance.LoadNamedComponents(components)
			return
		},
	})
}

func init() {
	MustRegisterContainerImport(compcont.DefaultFactoryRegistry)
}
