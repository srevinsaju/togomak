package ops

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path"

	"github.com/srevinsaju/buildsys/pkg/context"
	"github.com/srevinsaju/buildsys/pkg/schema"
)

func RunStage(ctx *context.Context, stage schema.StageConfig) {
	stageCtx := ctx.AddChild("stage", stage.Id)

	// show some user friendly output on the details of the stage about to be run 
	if stage.Name != "" && stage.Description != "" {
		stageCtx.Logger.Infof("Starting stage: ", stage.Name, " (", stage.Description, ")")
	} else if stage.Name != "" {
		stageCtx.Logger.Info("Starting stage: ", stage.Name)
	} else if stage.Description != "" {
		stageCtx.Logger.Info("Starting stage: ", stage.Description)
	} else {
		stageCtx.Logger.Info("Starting stage: ", stage.Id)
	}

	var err error

	if stage.Script != "" {
		stageCtx.Logger.Debug("Preparing script")
		if stage.Plugin != "" {
			stageCtx.Logger.Fatal("Cannot use both script and plugin")
		}

		tempTargetRunDir := path.Join(ctx.TempDir, stage.Id)
		targetRunPath := path.Join(tempTargetRunDir, "run.sh")
		stageCtx.Logger.Debug("Writing script to ", targetRunPath)
		err = os.MkdirAll(tempTargetRunDir, 0755)
		if err != nil {
			stageCtx.Logger.Fatal(err)
		}

		err = ioutil.WriteFile(targetRunPath, []byte(stage.Script), 0755)
		if err != nil {
			stageCtx.Logger.Fatal(err)
		}

		if stage.Container != "" {
			stageCtx.Logger.Debug("Preparing container")
			// the user requires a specific container to run on

			cmd := exec.Command("podman", "run", "--rm", "--entrypoint=sh", "-v", tempTargetRunDir+":/workspace:Z", stage.Container, "-c", "/workspace/run.sh")
			stageCtx.Logger.Debug("Running ", cmd.String())
			cmd.Stdout = stageCtx.Logger.Writer()
			cmd.Stderr = stageCtx.Logger.Writer()
			err = cmd.Run()
			if err != nil {
				stageCtx.Logger.Fatal(err)
			}
		} else {
			cmd := exec.Command("sh", "-c", targetRunPath)
			cmd.Stdout = stageCtx.Logger.Writer()
			cmd.Stderr = stageCtx.Logger.Writer()
			err = cmd.Run()
			if err != nil {
				stageCtx.Logger.Fatal(err)
			}

		}
	}

}
