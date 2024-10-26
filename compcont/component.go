package compcont

import (
	"regexp"
	"slices"
)

type ComponentType string

type ComponentName string

var componentNameRegexp = regexp.MustCompile("^[a-zA-Z_][a-zA-Z0-9_]*$")

func (n ComponentName) Validate() bool {
	return componentNameRegexp.Match([]byte(n))
}

type ComponentConfig struct {
	Type   ComponentType   `json:"type" yaml:"type"`     // 组件类型
	Refer  ComponentName   `json:"refer" yaml:"refer"`   // 来自其他组件的引用
	Deps   []ComponentName `json:"deps" yaml:"deps"`     // 构造该组件需要依赖的其他组件名称
	Config any             `json:"config" yaml:"config"` // 组件的自身配置
}

// 组件的结构
type Component struct {
	Type         ComponentType
	Dependencies map[ComponentName]struct{}
	Instance     any
}

type Context struct {
	Container IComponentContainer
	Name      ComponentName
}

func (c *Context) GetAbsolutePath() (path []ComponentName) {
	path = append(path, c.Name)
	currentNode := c.Container
	for {
		// 非根节点才加入path
		if n := currentNode.GetSelfComponentName(); n != "" {
			path = append(path, n)
		}
		parent := currentNode.GetParent()
		if parent == nil {
			break
		}
		currentNode = parent
	}
	slices.Reverse(path)
	return
}

// 一个组件工厂
type IComponentFactory interface {
	Type() ComponentType // 组件唯一类型名称
	// 组件创建器，这里并没有明确config应该到底是什么类型，可以放到具体实现上既可以是map也可以是struct
	CreateInstance(ctx Context, config any) (instance any, err error)
	DestroyInstance(ctx Context, instance any) (err error) // 组件销毁器
}
