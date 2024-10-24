package container

import (
	"fmt"
	"slices"

	"github.com/go-compcont/compcont/compcont"
)

// 拓扑排序
func topologicalSort(cfgMap map[compcont.ComponentName]compcont.ComponentConfig) ([]compcont.ComponentName, error) {
	// 计算每个节点的入度
	inDegree := make(map[compcont.ComponentName]int)
	for name := range cfgMap {
		inDegree[name] = 0
	}
	for name, cfg := range cfgMap {
		for _, dep := range cfg.Deps {
			if _, ok := cfgMap[dep]; !ok {
				return nil, fmt.Errorf("component config error, %w, dependency %s not found for component %s", compcont.ErrComponentDependencyNotFound, dep, name)
			}
			inDegree[dep]++
		}
	}

	// 初始化队列，将所有入度为 0 的节点加入队列
	queue := []compcont.ComponentName{}
	for name, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, name)
		}
	}

	// 拓扑排序
	result := []compcont.ComponentName{}
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		result = append(result, node)

		for _, dep := range cfgMap[node].Deps {
			inDegree[dep]--
			if inDegree[dep] == 0 {
				queue = append(queue, dep)
			}
		}
	}

	// 检查是否有环
	if len(result) != len(cfgMap) {
		return nil, compcont.ErrCircularDependency
	}

	slices.Reverse(result)
	return result, nil
}
