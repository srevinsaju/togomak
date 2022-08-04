package main

import (
	"fmt"
	"os"
	"os/exec"

	log "github.com/sirupsen/logrus"
	hclog "github.com/hashicorp/go-hclog"
	
	"github.com/hashicorp/go-plugin"
	"github.com/srevinsaju/buildsys/pkg/schema"
)

func init() {
	// Log as JSON instead of the default ASCII formatter.
	// log.SetFormatter(&log.JSONFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	log.SetLevel(log.TraceLevel)
}

func main() {

	ctxLog := log.WithField("context", "main")

	// Create an hclog.Logger
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "provider",
		Output: os.Stdout,
		Level:  hclog.Warn,
	})
	
	// We're a host! Start by launching the plugin process.
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: handshakeConfig,
		Plugins:         pluginMap,
		Cmd:             exec.Command("../../plugins/git/git"),
		Logger:          logger,
	})
	defer client.Kill()

	// Connect via RPC
	rpcClient, err := client.Client()
	if err != nil {
		ctxLog.Fatal(err)
	}

	// Request the plugin
	raw, err := rpcClient.Dispense("provider")
	if err != nil {
		ctxLog.Fatal(err)
	}

	// We should have a Greeter now! This feels like a normal interface
	// implementation but is in fact over an RPC connection.
	provider := raw.(schema.Stage)
	fmt.Println(provider.Name())
	provider.GatherInfo()
	ctxLog.Info("Gathered info")
	ctxLog.Info(provider.GetContext())
	
}

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
	"provider": &schema.StagePlugin{},
}
