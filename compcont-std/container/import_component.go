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
	FromFile     string                                            `ccf:"from_file"`
	ExportMapper map[compcont.ComponentName]compcont.ComponentName `ccf:"export_mapper"`
}

func MustRegisterContainerImport(r compcont.IFactoryRegistry) {
	r.Register(&compcont.TypedSimpleComponentFactory[ContainerImportConfig, compcont.IComponentContainer]{
		TypeName: ContainerImportType,
		CreateInstanceFunc: func(container compcont.IComponentContainer, config ContainerImportConfig) (instance compcont.IComponentContainer, err error) {
			instance = NewComponentContainer(WithFactoryRegistry(container.FactoryRegistry()))
			var bs []byte
			bs, err = os.ReadFile(config.FromFile)
			if err != nil {
				return
			}

			components := make(map[compcont.ComponentName]compcont.ComponentConfig)
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

			for inParent, inChild := range config.ExportMapper {
				var comp compcont.Component
				comp, err = instance.GetComponent(inChild)
				if err != nil {
					return
				}
				err = instance.PutComponent(inParent, comp)
				if err != nil {
					return
				}
			}
			return
		},
	})
}

func init() {
	MustRegisterContainerImport(compcont.DefaultFactoryRegistry)
}
