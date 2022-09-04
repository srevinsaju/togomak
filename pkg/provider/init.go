package provider

import (
	"fmt"
	"github.com/spf13/afero"
	"github.com/srevinsaju/togomak/pkg/meta"
	"github.com/srevinsaju/togomak/pkg/x"
	"os"
	"os/exec"
	"path/filepath"

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
	MagicCookieValue: meta.AppName,
}

// pluginMap is the map of plugins we can dispense.
var pluginMap = map[string]plugin.Plugin{
	"stage":    &schema.StagePlugin{},
	"data":     &schema.StagePlugin{},
	"provider": &schema.StagePlugin{},
}

var providers map[string]schema.Provider

func initProvider(ctx *context.Context, p schema.ProviderConfig) schema.Provider {
	ctx.Logger.Debugf("Loading provider %s", p.Name())
	// Create an hclog.Logger
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "provider",
		Output: os.Stdout,
		Level:  hclog.Warn,
	})
	if providers == nil {
		providers = make(map[string]schema.Provider)
	}
	if p.Path == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			panic(err)
		}
		ctx.Logger.Debugf("Searching under %s/plugins, %s/%s/plugins dir", meta.BuildDirPrefix, homeDir, meta.BuildDirPrefix)

		// first check if the current directory has a .togomak.plugins directory
		// and load plugins from there, else check home directory.
		cwd, err := os.Getwd()
		x.Must(err)

		cwdPluginDir := filepath.Join(cwd, meta.BuildDirPrefix, "plugins", fmt.Sprintf("togomak-provider-%s", p.Name()))
		exists, err := afero.Exists(afero.OsFs{}, cwdPluginDir)
		ctx.Logger.Debugf("Checking if %s exists", cwdPluginDir)
		if err != nil || !exists {
			ctx.Logger.Debugf("Failed loading provider %s from %s: %s", p.Name(), cwdPluginDir, err)
			togomakPluginDir := filepath.Join(homeDir, meta.BuildDirPrefix, "plugins", fmt.Sprintf("togomak-provider-%s", p.Name()))
			ctx.Logger.Debugf("Checking if %s exists", togomakPluginDir)
			exists, err := afero.Exists(afero.OsFs{}, togomakPluginDir)
			if err != nil || !exists {
				ctx.Logger.Warnf("Failed loading provider %s from %s: %s", p.Name(), togomakPluginDir, err)
				return schema.Provider{
					Config: p,
				}
			}
			ctx.Logger.Debugf("Found %s", togomakPluginDir)
			p.Path = togomakPluginDir
		} else {
			ctx.Logger.Debugf("Found %s", cwdPluginDir)
			p.Path = cwdPluginDir
		}
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
	}
	providers[p.Name()] = provider
	ctx.Logger.Trace("providers", providers)
	return provider

}

// Get returns a provider by the name of the provider
// at a time, there can only be one provider by name. There can be multiple providers
// with different ids on same ID.
func Get(ctx *context.Context, p schema.ProviderConfig) schema.Provider {
	if v, ok := providers[p.Name()]; ok {
		return v
	}
	return initProvider(ctx, p)
}

func Destroy(ctx *context.Context, p schema.ProviderConfig) {
	v, ok := providers[p.Name()]
	ctx.Logger.Tracef("Currently loaded providers are %v", providers)
	ctx.Logger.Tracef("Unloading provider %s", p.Name())

	if !ok {
		ctx.Logger.Warnf("Provider %s is not loaded", p.Name())
		panic("provider is not loaded on to memory yet")
	}

	// TODO: destroy the provider
	if p.Path != "" {
		defer v.Client.Kill()
	}

	delete(providers, p.Name())
}
