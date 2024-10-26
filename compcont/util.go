package compcont

import (
	"fmt"
	"slices"
)

type set[T comparable] map[T]struct{}

// 拓扑排序
func topologicalSort(cfgMap map[ComponentName]set[ComponentName]) ([]ComponentName, error) {
	// 计算每个节点的入度
	inDegree := make(map[ComponentName]int)
	{
		for name := range cfgMap {
			inDegree[name] = 0
		}
		for name, deps := range cfgMap {
			for dep := range deps {
				if _, ok := cfgMap[dep]; !ok { // 顺便校验下是否存在不存在的引用关系
					return nil, fmt.Errorf("component config error, %w, dependency %s not found for component %s", ErrComponentDependencyNotFound, dep, name)
				}
				inDegree[dep]++
			}
		}
	}

	// 初始化队列，将所有入度为 0 的节点加入队列
	var queue []ComponentName
	for name, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, name)
		}
	}

	// 拓扑排序
	var result []ComponentName
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		result = append(result, node)

		for dep := range cfgMap[node] {
			inDegree[dep]--
			if inDegree[dep] == 0 {
				queue = append(queue, dep)
			}
		}
	}

	// 检查是否有环
	if len(result) != len(cfgMap) {
		return nil, ErrCircularDependency
	}

	slices.Reverse(result)
	return result, nil
}

// 从当前节点定位一个组件的上下文
func find(currentNode IComponentContainer, findPath []ComponentName, absolute bool) (ctx Context, err error) {
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

	var component Component
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
			ctx = component.Context
			return
		}

		// 还要继续向后寻找，如果下一个要寻找的节点不是容器，则直接报错
		if _, ok := component.Instance.(IComponentContainer); !ok {
			err = fmt.Errorf("refer path error, %s is not a container", partName)
			return
		}
	}

	ctx = component.Context
	return
}
