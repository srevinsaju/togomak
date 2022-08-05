package main

import (
	"os"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/srevinsaju/togomak/pkg/schema"
)

// Here is a real implementation of Stage
type StageGit struct {
	logger  hclog.Logger
	context schema.Context
}

// handshakeConfigs are used to just do a basic handshake between
// a plugin and host. If the handshake fails, a user friendly error is shown.
// This prevents users from executing bad plugins or executing a plugin
// directory. It is a UX feature, not a security feature.
var handshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "BASIC_PLUGIN",
	MagicCookieValue: "togomak",
}

func main() {
	logger := hclog.New(&hclog.LoggerOptions{
		Level:      hclog.DefaultLevel,
		Output:     os.Stderr,
		JSONFormat: true,
		Color:      hclog.ForceColor,
	})

	git := &StageGit{
		//logger: logger,
		context: schema.Context{
			Data: map[string]string{},
			//Mutex: &sync.Mutex{},
		},
	}
	// pluginMap is the map of plugins we can dispense.
	var pluginMap = map[string]plugin.Plugin{
		"data": &schema.StagePlugin{Impl: git},
	}

	logger.Debug("message from plugin", "go", "bar")

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: handshakeConfig,
		Plugins:         pluginMap,
	})
}
