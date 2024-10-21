package compcont

import (
	"fmt"
	"slices"
	"time"

	"github.com/mitchellh/mapstructure"
)

func decodeMapConfig[C any](mapConfig map[string]any, structureConfig *C) (err error) {
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName:     "ccf",
		ErrorUnused: true,            // 配置文件如果多余出未使用的字段，则报错
		ZeroFields:  true,            // decode前对传入的结构体清零
		Result:      structureConfig, // 目标结构体
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),     // 自动解析duration
			mapstructure.StringToTimeHookFunc(time.RFC3339), // 自动解析时间
		),
	})
	if err != nil {
		return
	}
	err = decoder.Decode(mapConfig)
	if err != nil {
		return
	}
	return
}

func topologicalSort(cfgMap map[ComponentName]ComponentConfig) ([]ComponentName, error) {
	// 计算每个节点的入度
	inDegree := make(map[ComponentName]int)
	for name := range cfgMap {
		inDegree[name] = 0
	}
	for name, cfg := range cfgMap {
		for _, dep := range cfg.Deps {
			if _, ok := cfgMap[dep]; !ok {
				return nil, fmt.Errorf("component config error, %w, dependency %s not found for component %s", ErrComponentDependencyNotFound, dep, name)
			}
			inDegree[dep]++
		}
	}

	// 初始化队列，将所有入度为 0 的节点加入队列
	queue := []ComponentName{}
	for name, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, name)
		}
	}

	// 拓扑排序
	result := []ComponentName{}
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
		return nil, ErrCircularDependency
	}

	slices.Reverse(result)
	return result, nil
}
