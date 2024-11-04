package restyprovider

import (
	"context"
	"errors"
	"fmt"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
	"github.com/go-compcont/compcont/compcont"
	"github.com/go-compcont/compcont/compcont-std/compcontzap"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

type RuleConfig struct {
	Match         string                                            `ccf:"match"`
	RestyProvider compcont.TypedComponentConfig[any, RestyProvider] `ccf:"resty_provider"`
}

type Rule struct {
	source   string
	ruleExpr *vm.Program
	RestyProvider
}

type env struct {
	Host string   `expr:"host"`
	Path string   `expr:"path"`
	Tags []string `expr:"tags"`
}

func NewRule(cc compcont.IComponentContainer, cfg RuleConfig) (rule Rule, err error) {
	ruleExpr, err := expr.Compile(cfg.Match, expr.Env(env{}), expr.AsBool())
	if err != nil {
		return
	}
	restyProviderComp, err := cfg.RestyProvider.LoadComponent(cc)
	if err != nil {
		return
	}
	rule = Rule{
		RestyProvider: restyProviderComp.Instance,
		ruleExpr:      ruleExpr,
		source:        cfg.Match,
	}
	return
}

func (r *Rule) Match(opt options) (ok bool, err error) {
	var env env
	if opt.url != nil {
		env.Host = opt.url.Host
		env.Path = opt.url.Path
		env.Tags = opt.tags
	}
	ret, err := expr.Run(r.ruleExpr, env)
	if err != nil {
		return
	}
	if v, ok1 := ret.(bool); ok1 {
		ok = v
		return
	}
	err = fmt.Errorf("unexpected expr result: %+v, expected a bool value", ret)
	return
}

type RuleProviderConfig struct {
	DefaultProvider compcont.TypedComponentConfig[any, RestyProvider] `ccf:"default_provider"`
	Rules           []RuleConfig                                      `ccf:"rules"`
}

type ruleProviderImpl struct {
	defaultProvider RestyProvider
	rules           []Rule
}

func newRuleProviderImpl(cc compcont.IComponentContainer, cfg RuleProviderConfig) (c RestyProvider, err error) {
	defaultProviderComp, err := cfg.DefaultProvider.LoadComponent(cc)
	if err != nil {
		return
	}

	var rules []Rule
	for _, ruleCfg := range cfg.Rules {
		rule, err := NewRule(cc, ruleCfg)
		if err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}
	c = &ruleProviderImpl{
		defaultProvider: defaultProviderComp.Instance,
		rules:           rules,
	}
	return
}

func (c *ruleProviderImpl) GetResty(opts ...OptionsFunc) (cli *resty.Client, err error) {
	var opt options
	for _, fn := range opts {
		fn(&opt)
	}

	if opt.ctx == nil {
		opt.ctx = context.Background()
	}

	logger := compcontzap.FromContext(opt.ctx)
	for _, rule := range c.rules {
		ok, err := rule.Match(opt)
		if err != nil {
			logger.Error("rule match error", zap.Error(err))
			return nil, err
		}
		if ok {
			logger.Debug("match success", zap.String("rule", rule.source))
			return rule.GetResty(opts...)
		}
	}
	if c.defaultProvider == nil {
		logger.Error("no avaliable resty client")
		return nil, errors.New("no avaliable resty client")
	}
	logger.Debug("no match any resty in rules, use default")
	return c.defaultProvider.GetResty(opts...)
}
