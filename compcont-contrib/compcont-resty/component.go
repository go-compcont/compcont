package restyprovider

import "github.com/go-compcont/compcont/compcont"

const SimpleTypeID compcont.ComponentTypeID = "contrib.resty-provider-simple"

var simpleFactory compcont.IComponentFactory = &compcont.TypedSimpleComponentFactory[SimpleProviderConfig, RestyProvider]{
	TypeID: SimpleTypeID,
	CreateInstanceFunc: func(ctx compcont.Context, config SimpleProviderConfig) (instance RestyProvider, err error) {
		err = config.checkAndFillDefault()
		if err != nil {
			return
		}
		return newSimpleProviderImpl(ctx.Container, config)
	},
}

const RuleTypeID compcont.ComponentTypeID = "contrib.resty-provider-rule"

var ruleFactory compcont.IComponentFactory = &compcont.TypedSimpleComponentFactory[RuleProviderConfig, RestyProvider]{
	TypeID: RuleTypeID,
	CreateInstanceFunc: func(ctx compcont.Context, config RuleProviderConfig) (instance RestyProvider, err error) {
		return newRuleProviderImpl(ctx.Container, config)
	},
}

func MustRegister(registry compcont.IFactoryRegistry) {
	registry.Register(simpleFactory)
	registry.Register(ruleFactory)
}

func init() {
	MustRegister(compcont.DefaultFactoryRegistry)
}
