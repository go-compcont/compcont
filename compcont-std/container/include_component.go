package container

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/go-compcont/compcont/compcont"
	"gopkg.in/yaml.v3"
)

const ImportContainerType compcont.ComponentType = "std.import-container"

type ImportFileConfig map[string]compcont.ComponentConfig

type ImportContainerConfig struct {
	Import string `ccf:"import"`
}

func MustRegisterImportContainer(r compcont.IFactoryRegistry) {
	r.Register(&compcont.TypedSimpleComponentFactory[ImportContainerConfig, compcont.IComponentContainer]{
		TypeName: ImportContainerType,
		CreateInstanceFunc: func(container compcont.IComponentContainer, config ImportContainerConfig) (instance compcont.IComponentContainer, err error) {
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
	MustRegisterImportContainer(compcont.DefaultFactoryRegistry)
}
