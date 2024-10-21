package compcont

import (
	"time"

	"github.com/mitchellh/mapstructure"
)

func decodeMapConfig[C any](rawCfg map[string]any, cfg *C) (err error) {
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName:     "ccf",
		ErrorUnused: true, // 配置文件如果多余出未使用的字段，则报错
		ZeroFields:  true, // decode前对传入的结构体清零
		Result:      cfg,  // 目标结构体
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),     // 自动解析duration
			mapstructure.StringToTimeHookFunc(time.RFC3339), // 自动解析时间
		),
	})
	if err != nil {
		return
	}
	err = decoder.Decode(rawCfg)
	if err != nil {
		return
	}
	return
}
