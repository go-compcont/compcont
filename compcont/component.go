package compcont

import (
	"regexp"
	"slices"
)

type ComponentTypeID string

type ComponentName string

var componentNameRegexp = regexp.MustCompile("^[a-zA-Z_][a-zA-Z0-9_]*$")

func (n ComponentName) Validate() bool {
	return componentNameRegexp.Match([]byte(n))
}

type ComponentConfig struct {
	Name   ComponentName   `json:"name" yaml:"name"`     // 组件名称，不填为空值，即匿名组件
	Type   ComponentTypeID `json:"type" yaml:"type"`     // 组件类型
	Refer  string          `json:"refer" yaml:"refer"`   // 来自其他组件的引用
	Deps   []ComponentName `json:"deps" yaml:"deps"`     // 构造该组件需要依赖的其他组件名称
	Config any             `json:"config" yaml:"config"` // 组件的自身配置
}

// 运行时的组件的结构
type Component struct {
	Context  Context // 运行时一个组件必然存在一个Context，且不可变，这里使用值类型
	Instance any
}

// 构造组件时使用的上下文环境结构
type Context struct {
	Container IComponentContainer // 当前组件所在容器
	Config    ComponentConfig     // 组件配置
	Mount     *Component          // 组件实例有可能不存在
}

func (c *Context) FindRoot() Context {
	currentNode := c.Container
	for {
		parent := currentNode.GetParent()
		if parent == nil {
			break
		}
		currentNode = parent
	}
	return currentNode.GetContext()
}

func (c *Context) GetAbsolutePath() (path []ComponentName) {
	path = append(path, c.Config.Name)
	currentNode := c.Container
	for {
		// 非根节点才加入path
		parent := currentNode.GetParent()
		if parent == nil {
			break
		}
		path = append(path, currentNode.GetContext().Config.Name)
		currentNode = parent
	}
	slices.Reverse(path)
	return
}

// 一个组件工厂
type IComponentFactory interface {
	Type() ComponentTypeID // 组件唯一类型名称
	// 组件创建器，这里并没有明确config应该到底是什么类型，可以放到具体实现上既可以是map也可以是struct
	CreateInstance(ctx Context, config any) (instance any, err error)
	DestroyInstance(ctx Context, instance any) (err error) // 组件销毁器
}
