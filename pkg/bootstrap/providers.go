package bootstrap

import (
	"github.com/srevinsaju/togomak/pkg/context"
	"github.com/srevinsaju/togomak/pkg/provider"
	"github.com/srevinsaju/togomak/pkg/schema"
	"github.com/srevinsaju/togomak/pkg/x"
)

type InternalProviders map[string]schema.Provider

func Providers(ctx *context.Context, data schema.SchemaConfig) InternalProviders {
	var providers InternalProviders
	providers = make(InternalProviders)
	providerLog := ctx.Logger.WithField("context", "providers")
	providerLog.Debug("Bootstrapping providers")

	for _, p := range data.Providers {

		if _, ok := providers[p.ID()]; !ok {

			providerCtx := ctx.AddChild("provider", p.ID())
			providers[p.ID()] = provider.Get(providerCtx, p)

		} else {
			providerLog.Fatal("Duplicate provider ID: " + p.ID())
		}
	}
	return providers
}

func (providers InternalProviders) Get(id string) schema.Provider {
	return providers[id]
}

func (providers InternalProviders) SetContext(ctx *context.Context, data schema.SchemaConfig) {
	// providerLog := ctx.Logger.WithField("context", "providers")

	for _, p := range providers {
		providerCtx := ctx.AddChild("provider", p.Config.ID())
		providerCtx.DataMutex.Lock()
		providerCtx.Data["params"] = p.Config.Data
		providerCtx.DataMutex.Unlock()
		providerCtx.Logger.Debug("Sending default parameters")
		x.Must(p.Provider.SetContext(schema.Context{Data: p.Config.Data}))
	}
}

func (providers InternalProviders) GatherInfo(ctx *context.Context) {
	providerLog := ctx.Logger.WithField("context", "providers")

	for _, p := range providers {
		providerCtx := ctx.AddChild("provider", p.Config.ID())
		if p.Config.Path == "" {
			// TODO: fixme
			continue
		}
		providerCtx.Logger.Tracef("Requesting information from provider")

		err := p.Provider.GatherInfo()
		if err.IsErr {
			providerCtx.Logger.Fatal(err.Err)
		}
		for k, v := range p.Provider.GetContext().Data {

			providerCtx.Logger.Debugf("Received context from provider %s: %v", k, v)
			providerCtx.DataMutex.Lock()
			providerCtx.Data[k] = v
			providerCtx.DataMutex.Unlock()
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
