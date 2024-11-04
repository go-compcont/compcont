package restyprovider

import (
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/go-compcont/compcont/compcont"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

func TestRuleProvider(t *testing.T) {
	compcont.DefaultFactoryRegistry.Register(&compcont.TypedSimpleComponentFactory[string, RestyProvider]{
		TypeID: "mock.resty",
		CreateInstanceFunc: func(ctx compcont.Context, config string) (instance RestyProvider, err error) {
			comp := GetRestyFunc(func(opts ...OptionsFunc) (*resty.Client, error) {
				return resty.New().SetBaseURL(fmt.Sprint(config, time.Now().UnixMicro())), nil
			})
			instance = comp
			return
		},
	})

	cc := compcont.NewComponentContainer()

	p, err := newRuleProviderImpl(cc, RuleProviderConfig{
		DefaultProvider: compcont.TypedComponentConfig[any, RestyProvider]{
			Type:   "mock.resty",
			Config: "test-default",
		},
		Rules: []RuleConfig{
			{
				Match: `hasSuffix(path, "test")`,
				RestyProvider: compcont.TypedComponentConfig[any, RestyProvider]{
					Type:   "mock.resty",
					Config: "test-1",
				},
			},
		},
	})
	assert.NoError(t, err)
	cli, err := p.GetResty(func(o *options) {
		o.url, err = url.Parse("http://127.0.0.1/test")
		assert.NoError(t, err)
	})
	assert.NoError(t, err)
	assert.Contains(t, cli.BaseURL, "test-1")

	cli, err = p.GetResty(func(o *options) {
		o.url, err = url.Parse("http://127.0.0.1/test1")
		assert.NoError(t, err)
	})
	assert.NoError(t, err)
	assert.Contains(t, cli.BaseURL, "test-default")
}
