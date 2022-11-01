package runner

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/srevinsaju/togomak/pkg/bootstrap"
	"github.com/srevinsaju/togomak/pkg/config"
	"github.com/srevinsaju/togomak/pkg/context"
	"github.com/srevinsaju/togomak/pkg/meta"
	"github.com/srevinsaju/togomak/pkg/state"
	"github.com/srevinsaju/togomak/pkg/templating"
	"github.com/srevinsaju/togomak/pkg/ui"
	"os"
	"time"
)

const SupportedCiConfigVersion = 1
const StateWorkspace = "state_workspace"

func List(cfg config.Config) {

	/// create context
	ctx := &context.Context{
		Logger: log.WithFields(log.Fields{}),
		Data:   context.Data{},
	}
	/// load config
	data := bootstrap.Config(ctx, &cfg)
	for _, v := range data.Stages {
		fmt.Println(v.Id)
	}
}

func Orchestrator(cfg config.Config) {
	orchestratorStartTime := time.Now()

	owd, _ := os.Getwd()

	/// create context
	ctx := &context.Context{
		Logger: log.WithFields(log.Fields{}),
		Data: context.Data{
			// some default functions
			"owd": owd,
			"env": templating.Env,
		},
	}

	ctx.Logger.Debugf("Starting %s", meta.AppName)

	/// load config
	data := bootstrap.Config(ctx, &cfg)

	/// change working directory to the directory of the config file
	bootstrap.Chdir(ctx, cfg, data)

	/// create temporary working directory
	bootstrap.TempDir(ctx)
	defer bootstrap.SafeDeleteTempDir(ctx)

	stateUrl := data.State.URL
	if stateUrl == "" {
		stateUrl = fmt.Sprintf("file://%s", meta.BuildDirPrefix)
	}

	ctx.Data["default_state_manager"] = bootstrap.LoadStateBackend(ctx, stateUrl)
	if data.State.Workspace == "" {
		data.State.Workspace = meta.DefaultWorkspaceType
	}
	ctx.Data[state.WorkspaceDataKey] = data.State.Workspace

	/// get the parameters
	bootstrap.Params(ctx, data, cfg.NoInteractive)

	/// override parameters from the command line, cfg object
	bootstrap.OverrideParams(ctx, cfg)

	/// run initial validation
	bootstrap.StageValidate(ctx, data)

	/// expand sources
	bootstrap.ExpandSources(ctx, &data)

	/// generate dependency graph
	bootstrap.Graph(ctx, data)

	/// load the providers
	providers := bootstrap.Providers(ctx, data)
	providers.SetContext(ctx, data)

	/// gather information from all providers
	providers.GatherInfo(ctx)
	defer providers.UnloadAll(ctx)

	// check if matrix is specified
	ctx.Logger.Debugf("Need to run stages: %v", cfg.RunStages)
	if data.Matrix != nil {
		ctx.IsMatrix = true
		bootstrap.MatrixRun(ctx, data, cfg)
	} else {
		bootstrap.SimpleRun(ctx, cfg, data)
	}

	if cfg.Summary == config.SummaryOn || cfg.FailLazy || cfg.Summary != config.SummaryOff && config.GetSummaryType(data.Options.Summary) == config.SummaryOn {
		if cfg.RunStages == nil {
			bootstrap.Summary(ctx)
		}
	}
	ctx.Logger.Info(ui.Grey(fmt.Sprintf("togomak completed in %s", time.Now().Sub(orchestratorStartTime))))
}
