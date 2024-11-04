package restyprovider

import (
	"crypto/tls"
	"math/rand/v2"
	"sync"

	"github.com/go-compcont/compcont/compcont"
	"github.com/go-compcont/compcont/compcont-std/reloading"
	"github.com/go-resty/resty/v2"
)

func statusCodeHitTest(statusCodeCfg []int, code int) bool {
	for _, codeCfg := range statusCodeCfg {
		if codeCfg < 10 {
			if code/100 == codeCfg {
				return true
			}
		} else {
			if code == codeCfg {
				return true
			}
		}
	}
	return false
}

type simpleProviderImpl struct {
	SimpleProviderConfig
	userAgentProvider reloading.IReloadingConfig[[]string]
	onceClient        *resty.Client
	once              sync.Once
}

func (c *simpleProviderImpl) GetResty(opts ...OptionsFunc) (cli *resty.Client, err error) {
	if !c.SimpleProviderConfig.Once {
		return c.getRestyNoOnce()
	}
	c.once.Do(func() {
		c.onceClient, err = c.getRestyNoOnce()
	})
	return c.onceClient, err
}

func newSimpleProviderImpl(cc compcont.IComponentContainer, cfg SimpleProviderConfig) (comp *simpleProviderImpl, err error) {
	comp = &simpleProviderImpl{
		SimpleProviderConfig: cfg,
	}
	if cfg.UserAgent.UserAgentListProvider != nil {
		comp.userAgentProvider, err = cfg.UserAgent.UserAgentListProvider.Build(cc)
		if err != nil {
			return
		}
	}
	return
}

func (c *simpleProviderImpl) getRestyNoOnce() (cli *resty.Client, err error) {
	cli = resty.New().
		SetDebug(c.Debug.Enabled).
		SetDebugBodyLimit(*c.Debug.BodySizeLimit).
		SetTLSClientConfig(&tls.Config{InsecureSkipVerify: c.TLS.InsecureSkipVerify}).
		SetTimeout(c.Timeout).
		SetRetryMaxWaitTime(*c.Retry.MaxWaitTime).
		SetRetryWaitTime(*c.Retry.WaitTime).
		SetRetryCount(*c.Retry.MaxCount).
		OnBeforeRequest(func(cli *resty.Client, r *resty.Request) error {
			if c.userAgentProvider != nil {
				userAgentList, err := c.userAgentProvider.LoadConfig(r.Context())
				if err != nil {
					return err
				}
				if len(userAgentList) > 0 {
					userAgent := userAgentList[rand.N(len(userAgentList))]
					r.Header.Set("User-Agent", userAgent)
				}
			}
			return nil
		})
	if c.Proxy.Enabled {
		cli.SetProxy(c.Proxy.ProxyAddr)
	} else {
		cli.RemoveProxy()
	}
	for _, retryCondition := range c.SimpleProviderConfig.Retry.Condition {
		cli.AddRetryCondition(func(r *resty.Response, err error) bool {
			if err != nil {
				return true
			}
			if statusCodeHitTest(retryCondition.StatusCode, r.StatusCode()) {
				return true
			}
			return false
		})
	}
	return
}
