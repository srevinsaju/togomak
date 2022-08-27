package bootstrap

import (
	"github.com/srevinsaju/togomak/pkg/context"
	"github.com/srevinsaju/togomak/pkg/provider"
	"github.com/srevinsaju/togomak/pkg/schema"
)

type InternalProviders map[string]schema.Provider

func Providers(ctx *context.Context, data schema.SchemaConfig) InternalProviders {
	var providers InternalProviders
	providers = make(InternalProviders)
	providerLog := ctx.Logger.WithField("context", "providers")
	providerLog.Debug("Bootstrapping providers")

	for _, p := range data.Providers {
		if _, ok := providers[p.Id]; !ok {
			providerCtx := ctx.AddChild("provider", p.Id)
			providers[p.Id] = provider.Get(providerCtx, p)

		} else {
			providerLog.Fatal("Duplicate provider ID: " + p.Id)
		}
	}
	return providers
}

func (providers InternalProviders) Get(id string) schema.Provider {
	return providers[id]
}

func (providers InternalProviders) GatherInfo(ctx *context.Context) {
	providerLog := ctx.Logger.WithField("context", "providers")

	for _, p := range providers {
		if p.Config.Path == "" {
			continue
		}
		providerLog.Tracef("Requesting information from provider %s", p.Config.Id)

		err := p.Provider.GatherInfo()
		if err != nil {
			p.Context.Logger.Fatal(err)
		}
		for k, v := range p.Provider.GetContext().Data {
			p.Context.Logger.Debugf("Received context from provider %s: %v", k, v)
			p.Context.Data[k] = v
		}
	}

	providerLog.Tracef("Context before build: %v", ctx.Data)
}

func (providers InternalProviders) UnloadAll(ctx *context.Context) {
	ctx.Logger.WithField("context", "providers").Debug("Unloading providers")
	for _, p := range providers {
		provider.Destroy(ctx, p.Config)
	}
}
