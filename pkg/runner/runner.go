package runner

import (
	"github.com/kendru/darwin/go/depgraph"
	cartesian "github.com/schwarmco/go-cartesian-product"
	log "github.com/sirupsen/logrus"
	"github.com/srevinsaju/togomak/pkg/config"
	"github.com/srevinsaju/togomak/pkg/context"
	"github.com/srevinsaju/togomak/pkg/meta"
	"github.com/srevinsaju/togomak/pkg/provider"
	"github.com/srevinsaju/togomak/pkg/schema"
	"github.com/srevinsaju/togomak/pkg/ui"

	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"path/filepath"
)

const SupportedCiConfigVersion = 1

func Runner(cfg config.Config) {

	/// create context
	ctx := &context.Context{
		Logger: log.WithFields(log.Fields{}),
		Data:   map[string]interface{}{},
	}
	ctx.Logger.Debugf("Starting %s", meta.AppName)

	ctx.Logger.Debugf("Reading config file %s", cfg.CiFile)
	yfile, err := ioutil.ReadFile(cfg.CiFile)
	if err != nil {
		ctx.Logger.Fatal(err)
	}

	data := schema.SchemaConfig{}
	err = yaml.Unmarshal(yfile, &data)
	if err != nil {
		ctx.Logger.Fatal(err)
	}

	ctx.Logger.Tracef("Checking version of %s config", meta.AppName)
	if data.Version != SupportedCiConfigVersion {

		ctx.Logger.Fatal("Unsupported version on togomak config")
	}

	ctx.Logger.Debugf("Need to run stages: %v", cfg.RunStages)

	if !data.Options.Chdir && cfg.ContextDir == "" {
		// change working directory to the directory of the config file
		cwd := filepath.Dir(cfg.CiFile)
		ctx.Logger.Debugf("Changing directory to %s", cwd)
		err = os.Chdir(cwd)
		if err != nil {
			ctx.Logger.Warn(err)
		}
	} else {
		ctx.Logger.Debugf("Changing directory to %s", cfg.ContextDir)
		err = os.Chdir(cfg.ContextDir)
		if err != nil {
			ctx.Logger.Fatal(err)
		}
	}

	ctx.Logger.Tracef("loaded data: %v", data)

	tempDir, err := os.MkdirTemp("", ".togomak")
	defer os.RemoveAll(tempDir)
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

	rootStage := schema.NewRootStage()
	// generate the dependency graph with topological sort
	graph := depgraph.New()
	for _, stage := range data.Stages {
		if len(stage.DependsOn) == 0 {
			// no depends on
			validateLog.Debugf("%s stage depends on %s stage", stage.Id, rootStage.Id)
			err = graph.DependOn(stage.Id, rootStage.Id)
			if err != nil {
				ctx.Logger.Warn("Error while creating the dependency tree", err)
			}
		}
		for _, dep := range stage.DependsOn {
			validateLog.Debugf("%s stage depends on %s stage", dep, stage.Id)
			err = graph.DependOn(stage.Id, dep)
			if err != nil {
				ctx.Logger.Warn("Error while creating the dependency tree", err)
			}

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

	// unload providers
	defer func() {
		ctx.Logger.Debug("Unloading providers")
		for _, p := range providers {
			provider.Destroy(ctx, p.Config)
		}
	}()

	// check if matrix is specified
	if data.Matrix != nil {

		matrixLogger := ctx.Logger
		var keys []string
		var s [][]interface{}
		for k, v := range data.Matrix {
			keys = append(keys, k)
			var ss []interface{}
			for _, vv := range v {
				ss = append(ss, vv)
			}
			s = append(s, ss)
		}

		matrixText := ui.Grey("matrix")
		for product := range cartesian.Iter(s...) {

			matrixLogger.Infof("[%s] %s %s build", ui.Plus, ui.SubStage, ui.Matrix)

			ctx.Data["matrix"] = map[string]string{}
			for i := range keys {
				matrixLogger.Infof("%s %s %s.%s=%s", ui.SubStage, ui.SubSubStage, matrixText, ui.Grey(keys[i]), product[i])
				ctx.Data["matrix"].(map[string]string)[keys[i]] = product[i].(string)
			}

			Run(ctx, cfg, data, graph)
		}

	} else {
		Run(ctx, cfg, data, graph)
	}

}

func contains(cfg config.Config, l string) bool {
	for _, s := range cfg.RunStages {
		if s == l {
			return true
		}
	}
	return false

}
