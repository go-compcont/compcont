package finder

import (
	"fmt"
	"strings"

	"github.com/go-compcont/compcont/compcont"
)

const TypeName compcont.ComponentType = "std.finder"

func find(currentNode compcont.IComponentContainer, findPath []compcont.ComponentName, absolute bool) (instance any, err error) {
	var component compcont.Component
	// 如果是绝对路径，将currentNode指针指向容器树的根节点
	if absolute {
		for {
			parent := currentNode.GetParent()
			if parent == nil {
				break
			}
			currentNode = parent
		}
	}

	for i, partName := range findPath {
		if partName == "." {
			continue
		}
		if partName == ".." {
			currentNode = currentNode.GetParent()
			continue
		}
		component, err = currentNode.GetComponent(partName)
		if err != nil {
			return
		}

		// 已经找到最后一个路径了，可以返回
		if i == len(findPath)-1 {
			instance = component.Instance
			return
		}

		// 还要继续向后寻找，如果下一个要寻找的节点不是容器，则直接报错
		if _, ok := component.Instance.(compcont.IComponentContainer); !ok {
			err = fmt.Errorf("refer path error, %s is not a container", partName)
			return
		}
	}

	instance = currentNode
	return
}

func MustRegister(r compcont.IFactoryRegistry) {
	r.Register(&compcont.TypedSimpleComponentFactory[string, any]{
		TypeName: TypeName,
		CreateInstanceFunc: func(ctx compcont.Context, path string) (instance any, err error) {
			if path != "" {
				parts := strings.Split(path, "/")
				absolute := false
				if parts[0] == "" { // 绝对路径
					absolute = true
					parts = parts[1:]
				}

				var findPath []compcont.ComponentName
				for _, p := range parts {
					findPath = append(findPath, compcont.ComponentName(p))
				}
				return find(ctx.Container, findPath, absolute)
			}
			err = fmt.Errorf("refer component arguments invalid")
			return
		},
	})
}

func init() {
	MustRegister(compcont.DefaultFactoryRegistry)
}
