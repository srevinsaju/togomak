package ops

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path"

	"github.com/flosch/pongo2/v6"

	"github.com/srevinsaju/togomak/pkg/context"
	"github.com/srevinsaju/togomak/pkg/schema"
)

func RunStage(stageCtx *context.Context, stage schema.StageConfig) {

	rootCtx := stageCtx.RootParent()
	// show some user friendly output on the details of the stage about to be run
	if stage.Name != "" && stage.Description != "" {
		stageCtx.Logger.Infof("Stage -> %s (%s)", stage.Name, stage.Description)
	} else if stage.Name != "" {
		stageCtx.Logger.Infof("Stage -> %s", stage.Name)
	} else if stage.Description != "" {
		stageCtx.Logger.Infof("Stage -> %s", stage.Description)
	} else {
		stageCtx.Logger.Infof("Stage -> %s", stage.Id)
	}

	var err error

	if stage.Script != "" && len(stage.Args) != 0 {
		// both script and args cannot be set simulatanously
		stageCtx.Logger.Fatal("Script and Args cannot be set simulatanously")
	}

	if stage.Script != "" {
		stageCtx.Logger.Debug("Preparing script")
		if stage.Plugin != "" {
			stageCtx.Logger.Fatal("Cannot use both script and plugin")
		}

		tempTargetRunDir := path.Join(stageCtx.TempDir, stage.Id)
		targetRunPath := path.Join(tempTargetRunDir, "run.sh")
		stageCtx.Logger.Debug("Writing script to ", targetRunPath)
		err = os.MkdirAll(tempTargetRunDir, 0755)
		if err != nil {
			stageCtx.Logger.Fatal(err)
		}

		tpl, err := pongo2.FromString(stage.Script)
		if err != nil {
			stageCtx.Logger.Fatal(err)
		}
		data, err := tpl.Execute(pongo2.Context(stageCtx.RootParent().Data))
		if err != nil {
			stageCtx.Logger.Fatal(err)
		}

		err = ioutil.WriteFile(targetRunPath, []byte(data), 0755)
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
	} else {
		// run the args
		newArgs := make([]string, len(stage.Args))

		for i, arg := range stage.Args {
			// render them with pongo
			tpl, err := pongo2.FromString(arg)
			if err != nil {
				stageCtx.Logger.Fatal("Cannot render args:", err)
			}
			parsed, err := tpl.Execute(rootCtx.Data)
			if err != nil {
				stageCtx.Logger.Fatal("Cannot render args:", err)
			}
			newArgs[i] = parsed

		}
	}
}
