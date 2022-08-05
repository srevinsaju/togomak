package provider

import (
	"os"
	"os/exec"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"

	"github.com/srevinsaju/togomak/pkg/context"
	"github.com/srevinsaju/togomak/pkg/schema"
)

// handshakeConfigs are used to just do a basic handshake between
// a plugin and host. If the handshake fails, a user friendly error is shown.
// This prevents users from executing bad plugins or executing a plugin
// directory. It is a UX feature, not a security feature.
var handshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "BASIC_PLUGIN",
	MagicCookieValue: "buildsys",
}

// pluginMap is the map of plugins we can dispense.
var pluginMap = map[string]plugin.Plugin{
	"stage":    &schema.StagePlugin{},
	"data":     &schema.StagePlugin{},
	"provider": &schema.StagePlugin{},
}

var providers map[string]schema.Provider

func initProvider(ctx *context.Context, p schema.ProviderConfig) schema.Provider {
	ctx.Logger.Debugf("Loading provider %s", p.Id)
	// Create an hclog.Logger
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "provider",
		Output: os.Stdout,
		Level:  hclog.Warn,
	})

	if p.Path == "" {
		if providers == nil {
			providers = make(map[string]schema.Provider)
		}
		provider := schema.Provider{
			Config:  p,
			Context: ctx,
		}
		providers[p.Id] = provider
		ctx.Logger.Debug("providers", providers)
		return provider

	}

	// We're a host! Start by launching the plugin process.
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: handshakeConfig,
		Plugins:         pluginMap,
		Cmd:             exec.Command(p.Path),
		Logger:          logger,
	})

	// Connect via RPC
	rpcClient, err := client.Client()
	if err != nil {
		ctx.Logger.Fatal(err)
	}

	// Request the plugin
	raw, err := rpcClient.Dispense("data")
	if err != nil {
		ctx.Logger.Fatal(err)
	}

	provider := schema.Provider{
		Config:   p,
		Client:   client,
		Provider: raw.(schema.Stage),
		Context:  ctx,
	}
	providers[p.Id] = provider
	ctx.Logger.Trace("providers", providers)
	return provider

}

func Get(ctx *context.Context, p schema.ProviderConfig) schema.Provider {
	if v, ok := providers[p.Id]; ok {
		return v
	}
	return initProvider(ctx, p)
}

func Destroy(ctx *context.Context, p schema.ProviderConfig) {
	v, ok := providers[p.Id]
	ctx.Logger.Tracef("Currently loaded providers are %s", providers)
	ctx.Logger.Tracef("Unloading provider %s", p.Id)

	if !ok {
		ctx.Logger.Warnf("Provider %s is not loaded", p.Id)
		panic("provider is not loaded on to memory yet")
	}

	// TODO: destroy the provider
	if p.Path != "" {
		defer v.Client.Kill()
	}

	delete(providers, p.Id)
}
