package bootstrap

import (
	"github.com/flosch/pongo2/v6"
	"github.com/srevinsaju/togomak/pkg/config"
	"github.com/srevinsaju/togomak/pkg/context"
	"github.com/srevinsaju/togomak/pkg/ops"
	"github.com/srevinsaju/togomak/pkg/schema"
	"strings"
	"sync"
)

func SimpleRun(ctx *context.Context, cfg config.Config, data schema.SchemaConfig) {
	rootStage := schema.NewRootStage()
	ctx.Logger.Debug("Sorting dependency tree")

	for _, layer := range ctx.Graph.TopoSortedLayers() {

		var wg sync.WaitGroup
		jobCount := 0
		// run the jobs
		for _, l := range layer {
			if l == rootStage.Id {
				continue
			}
			if len(cfg.RunStages) > 0 && !contains(cfg, l) {
				ctx.Logger.Debugf("Skipping stage %s", l)
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

			if strings.ToLower(strings.TrimSpace(condition)) == "false" && len(cfg.RunStages) == 0 {
				// the stage should not be executed
				// the stage will only not be executed if it has not been specified manually in the cli
				stageCtx.Logger.Info("Skipping stage")
				continue
			}

			wg.Add(1)
			jobCount++

			go func(l string) {
				defer wg.Done()
				ops.PrepareStage(ctx, stage)
				ops.RunStage(cfg, stageCtx, stage)
			}(l)

			if jobCount == cfg.JobsNumber {
				// wait until we only run the jobs as specified by the -j param
				wg.Wait()
				jobCount = 0
			}
			if cfg.DryRun {
				// we do not want to print the output of the dry run
				// concurrently since it will create mixed up output
				wg.Wait()
			}
		}

		wg.Wait()

	}

	ctx.Logger.Debug("All stages completed")
}
