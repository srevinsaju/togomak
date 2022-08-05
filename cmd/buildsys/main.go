package main

import (
	"io/ioutil"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"

	"os"

	"github.com/flosch/pongo2/v6"
	"github.com/kendru/darwin/go/depgraph"
	log "github.com/sirupsen/logrus"

	"github.com/srevinsaju/buildsys/pkg/context"
	"github.com/srevinsaju/buildsys/pkg/ops"
	"github.com/srevinsaju/buildsys/pkg/provider"
	"github.com/srevinsaju/buildsys/pkg/schema"
)

func init() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.TextFormatter{
		DisableTimestamp: true,
	})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	if os.Getenv("BUILDSYS_DEBUG") != "" {
		log.SetLevel(log.TraceLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

}

func main() {

	ctx := &context.Context{
		Logger: log.WithFields(log.Fields{}),
		Data:   map[string]interface{}{},
	}

	ctx.Logger.Debugf("Starting buildsys")
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
	validateLog.Trace("Validating YAML config")

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

	// load the providers
	providers := map[string]schema.Provider{}

	for _, p := range data.Providers {
		if _, ok := providers[p.Id]; !ok {
			providerCtx := ctx.AddChild("provider", p.Id)
			providers[p.Id] = provider.Get(providerCtx, p)

		} else {
			validateLog.Fatal("Duplicate provider ID: " + p.Id)
		}

	}

	// gather information from all providers
	for _, p := range providers {
		if p.Config.Path == "" {
			continue
		}
		ctx.Logger.Tracef("Requesting information from provider %s", p.Config.Id)

		err := p.Provider.GatherInfo()
		if err != nil {
			p.Context.Logger.Fatal(err)
		}
		for k, v := range p.Provider.GetContext().Data {
			p.Context.Logger.Debugf("Received context from provider %s: %v", k, v)
			p.Context.Data[k] = v
		}

	}

	ctx.Logger.Tracef("Context before build %v", ctx.Data)
	//ctx.Logger.Tracef("SHA is %s", ctx.Data["provider"].(map[string]interface{})["git"].(map[string]interface{})["sha"])

	// unload providers
	defer func() {
		ctx.Logger.Debug("Unloading providers")
		for _, p := range providers {
			provider.Destroy(ctx, p.Config)
		}
	}()

	ctx.Logger.Debug("Sorting dependency tree")

	for _, layer := range graph.TopoSortedLayers() {

		var wg sync.WaitGroup

		// run the jobs
		for _, l := range layer {
			if l == rootStage.Id {
				continue
			}
			stage := data.Stages.GetStageById(l)
			stageCtx := ctx.AddChild("stage", stage.Id)

			tpl, err := pongo2.FromString(stage.Condition)
			if err != nil {
				stageCtx.Logger.Fatal("Failed to parse condition", err)
			}
			condition, err := tpl.Execute(ctx.Data)
			if err != nil {
				stageCtx.Logger.Fatal("Failed to execute condition", err)
			}
			stageCtx.Logger.Debugf("condition towards running stage is %s", condition)

			if strings.ToLower(strings.TrimSpace(condition)) == "false" {
				// the stage should not be executed
				stageCtx.Logger.Info("Skipping stage")
				continue
			}

			wg.Add(1)
			go func(l string) {
				defer wg.Done()
				ops.RunStage(stageCtx, stage)
			}(l)
		}

		wg.Wait()

	}

	ctx.Logger.Debug("All stages completed")

}
