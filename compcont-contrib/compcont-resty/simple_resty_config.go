package restyprovider

import (
	"time"

	"github.com/go-compcont/compcont/compcont-std/reloading"
)

func fillDefaultInPtr[T any](pptr **T, val T) {
	if *pptr == nil {
		*pptr = &val
	}
}

type DebugConfig struct {
	Enabled       bool   `ccf:"enabled"`         // 启用debug日志
	BodySizeLimit *int64 `ccf:"body_size_limit"` // debug日志的body大小
}

func (c *DebugConfig) checkAndFillDefault() (err error) {
	fillDefaultInPtr(&c.BodySizeLimit, 2*1024)
	return
}

type RetryCondition struct {
	StatusCode []int `ccf:"status_code"` // 重试状态码
}

type RetryConfig struct {
	MaxCount    *int             `ccf:"max_count"`     // 最大重试次数
	WaitTime    *time.Duration   `ccf:"wait_time"`     // 重试等待时间
	MaxWaitTime *time.Duration   `ccf:"max_wait_time"` // 总最大重试等待时间
	Condition   []RetryCondition `ccf:"condition"`     // 重试条件，或关系
}

func (c *RetryConfig) checkAndFillDefault() (err error) {
	fillDefaultInPtr(&c.MaxCount, 3)
	fillDefaultInPtr(&c.WaitTime, time.Millisecond*100)
	fillDefaultInPtr(&c.MaxWaitTime, 2*time.Second)

	if len(c.Condition) == 0 {
		// 默认5xx错误需要重试
		c.Condition = []RetryCondition{{StatusCode: []int{5}}}
	}
	return
}

type TLSConfig struct {
	InsecureSkipVerify bool `ccf:"insecure_skip_verify"` // 允许不信任的HTTPS
}

type UserAgentConfig struct {
	UserAgentListProvider *reloading.ReloadingConfigConfig[[]string] `ccf:"user_agent_list_provider"`
}

type ProxyConfig struct {
	Enabled   bool   `ccf:"enabled"`    // 启用代理
	ProxyAddr string `ccf:"proxy_addr"` // 代理地址
}

type SimpleProviderConfig struct {
	Once      bool            `ccf:"once"` // 是否单例
	Debug     DebugConfig     `ccf:"debug"`
	Timeout   time.Duration   `ccf:"timeout"`
	Proxy     ProxyConfig     `ccf:"proxy"`
	TLS       TLSConfig       `ccf:"tls"`
	Retry     RetryConfig     `ccf:"retry"`
	UserAgent UserAgentConfig `ccf:"user_agent"`
}

func (c *SimpleProviderConfig) checkAndFillDefault() (err error) {
	if err = c.Debug.checkAndFillDefault(); err != nil {
		return
	}
	if err = c.Retry.checkAndFillDefault(); err != nil {
		return
	}
	if c.Timeout == 0 {
		c.Timeout = time.Minute
	}
	return
}
