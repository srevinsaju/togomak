package bootstrap

import (
	"errors"
	"fmt"
	"github.com/flosch/pongo2/v6"
	"github.com/gobwas/glob"
	"github.com/srevinsaju/togomak/pkg/config"
	"github.com/srevinsaju/togomak/pkg/context"
	"github.com/srevinsaju/togomak/pkg/ops"
	"github.com/srevinsaju/togomak/pkg/schema"
	"github.com/srevinsaju/togomak/pkg/ui"
	"io/fs"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

func SimpleRun(ctx *context.Context, cfg config.Config, data schema.SchemaConfig) {

	rootStage := schema.NewRootStage()
	ctx.Logger.Debug("Sorting dependency tree")

	for _, layer := range ctx.Graph.TopoSortedLayers() {

		var wg sync.WaitGroup
		jobCount := 0
		// run the jobs
		for _, l := range layer {
			jobStartTime := time.Now()
			if l == rootStage.Id {
				continue
			}
			if len(cfg.RunStages) > 0 && !contains(cfg, l) {
				ctx.Logger.Debugf("Skipping stage %s", l)
				continue
			}

			stage := data.Stages.GetStageById(l)
			stageCtx := ctx.AddChild("stage", stage.Id)

			// check if the stage need to be run
			state, stateManager := GetStateForStage(ctx, stage)

			targetStartTime := time.Now()
			targetIsUptoDate := true
			for _, t := range stage.Targets {
				if strings.HasPrefix(t, "./") {
					t = t[2:]
				}
				stageCtx.Logger.Tracef("Expanding glob target: %s", t)
				g, err := glob.Compile(t)
				if err != nil {
					stageCtx.Logger.Warnf("Provided glob expression '%s' may be incorrect. Please check the following error message for more details.", t)
					stageCtx.Logger.Fatal(err)
				}

				_ = filepath.Walk(".", func(f string, info fs.FileInfo, err error) error {
					if !g.Match(f) {
						return nil
					}
					stageCtx.Logger.Tracef("Checking if %s is up to date", f)
					if !state.IsTargetUpToDate(f) {
						stageCtx.Logger.Debugf("Target %s is not up to date", f)
						targetIsUptoDate = false
						return errors.New("target is not up to date")
					}
					return nil
				})

			}
			stageCtx.Logger.Tracef("target sync check took %s", time.Now().Sub(targetStartTime))

			if targetIsUptoDate && (stage.Targets != nil || (cfg.Force || cfg.RunAll)) {
				stageCtx.Logger.Infof(ui.Yellow(fmt.Sprintf("Stage %s is up to date", stage.Id)))
				continue
			}

			tpl, err := pongo2.FromString(stage.Condition)
			if err != nil {
				stageCtx.Logger.Fatal("Failed to parse condition", err)
			}
			condition, err := tpl.Execute(ctx.Data.AsMap())
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

			stageCtx.Logger.Tracef("stage condition check took %s", time.Now().Sub(jobStartTime))

			wg.Add(1)
			jobCount++

			go func(l string) {
				jobPreparationStartTime := time.Now()
				defer wg.Done()
				ops.PrepareStage(ctx, &stage)
				ops.RunStage(cfg, stageCtx, stage)
				stageCtx.Logger.Tracef("stage run took %s", time.Now().Sub(jobPreparationStartTime))
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

			UpdateStateForStage(ctx, stage, stateManager)
		}

		wg.Wait()

	}

	ctx.Logger.Debug("All stages completed")
}
