package runner

import (
	log "github.com/sirupsen/logrus"
	"github.com/srevinsaju/togomak/pkg/bootstrap"
	"github.com/srevinsaju/togomak/pkg/config"
	"github.com/srevinsaju/togomak/pkg/context"
	"github.com/srevinsaju/togomak/pkg/meta"
)

const SupportedCiConfigVersion = 1

func Orchestrator(cfg config.Config) {

	/// create context
	ctx := &context.Context{
		Logger: log.WithFields(log.Fields{}),
		Data:   map[string]interface{}{},
	}
	ctx.Logger.Debugf("Starting %s", meta.AppName)

	/// load config
	data := bootstrap.Config(ctx, &cfg)

	/// change working directory to the directory of the config file
	bootstrap.Chdir(ctx, cfg, data)

	/// create temporary working directory
	bootstrap.TempDir(ctx)
	defer bootstrap.SafeDeleteTempDir(ctx)

	/// run initial validation
	bootstrap.StageValidate(ctx, data)

	/// generate dependency graph
	bootstrap.Graph(ctx, data)

	/// load the providers
	providers := bootstrap.Providers(ctx, data)

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
		bootstrap.Summary(ctx)
	}
}
