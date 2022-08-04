package main

import (
	"io/ioutil"
	"sync"

	"gopkg.in/yaml.v3"

	"os"
	"os/exec"

	hclog "github.com/hashicorp/go-hclog"
	"github.com/kendru/darwin/go/depgraph"
	log "github.com/sirupsen/logrus"

	"github.com/hashicorp/go-plugin"
	"github.com/srevinsaju/buildsys/pkg/context"
	"github.com/srevinsaju/buildsys/pkg/ops"
	"github.com/srevinsaju/buildsys/pkg/schema"
)

func init() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.TextFormatter{PadLevelText: true})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	log.SetLevel(log.TraceLevel)
}

func main() {

	ctx := &context.Context{
		Logger: log.WithFields(log.Fields{}),
	}

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
		ctx.Logger.Fatal(err)
	}

	// Request the plugin
	raw, err := rpcClient.Dispense("data")
	if err != nil {
		ctx.Logger.Fatal(err)
	}

	// We should have a Greeter now! This feels like a normal interface
	// implementation but is in fact over an RPC connection.
	provider := raw.(schema.Stage)
	provider.GatherInfo()
	ctx.Logger.Info("Gathered info")
	ctx.Logger.Info(provider.GetContext())

	fileName := os.Args[len(os.Args)-1]
	yfile, err := ioutil.ReadFile(fileName)
	if err != nil {
		ctx.Logger.Fatal(err)
	}
	data := schema.SchemaConfig{}
	err = yaml.Unmarshal(yfile, &data)
	if err != nil {
		ctx.Logger.Fatal(err)
	}

	ctx.Logger.Debug("data", data)

	tempDir, err := os.MkdirTemp("", ".buildsys")
	if err != nil {
		ctx.Logger.Fatal(err)
	}
	ctx.TempDir = tempDir

	validateLog := log.WithFields(log.Fields{"context": "validate"})
	validateLog.Info("Validating YAML config")

	stages := map[string]string{}

	// check if duplicate ID is present
	for _, stage := range data.Stages {
		if stage.Id == "" {
			validateLog.Fatal("Stage ID is empty")
		}
		if _, ok := stages[stage.Id]; ok {
			validateLog.Fatal("Duplicate stage ID: " + stage.Id)
		}
		stages[stage.Id] = stage.Id
	}

	rootStage := schema.StageConfig{
		Id:     "root",
		Script: "echo Root Stage",
	}
	// generate the dependency graph with topological sort
	graph := depgraph.New()
	for _, stage := range data.Stages {
		if len(stage.DependsOn) == 0 {
			// no depends on
			validateLog.Debugf("%s stage depends on %s stage", stage.Id, rootStage.Id)
			graph.DependOn(stage.Id, rootStage.Id)
		}
		for _, dep := range stage.DependsOn {
			validateLog.Debugf("%s stage depends on %s stage", dep, stage.Id)
			graph.DependOn(stage.Id, dep)
		}

	}

	ctx.Logger.Info("Sorting dependency tree")

	for _, layer := range graph.TopoSortedLayers() {

		var wg sync.WaitGroup

		// run the jobs
		for _, l := range layer {
			if l == rootStage.Id {
				continue
			}
			stage := data.Stages.GetStageById(l)
			wg.Add(1)
			go func(l string) {
				defer wg.Done()
				ops.RunStage(ctx, stage)
			}(l)
		}

		wg.Wait()

	}

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
	"stage":    &schema.StagePlugin{},
	"data":     &schema.StagePlugin{},
	"provider": &schema.StagePlugin{},
}
