package ops

import (
	"fmt"
	"github.com/srevinsaju/togomak/pkg/config"
	"github.com/srevinsaju/togomak/pkg/ui"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/flosch/pongo2/v6"

	"github.com/srevinsaju/togomak/pkg/context"
	"github.com/srevinsaju/togomak/pkg/schema"
)

func PrepareStage(ctx *context.Context, stage schema.StageConfig) {
	// show some user-friendly output on the details of the stage about to be run
	if stage.Name != "" && stage.Description != "" {
		ctx.Logger.Infof("[%s] %s (%s)", ui.Plus, ui.Yellow(stage.Name), ui.Grey(stage.Description))
	} else if stage.Name != "" {
		ctx.Logger.Infof("[%s] %s", ui.Plus, ui.Yellow(stage.Name))
	} else if stage.Description != "" {
		ctx.Logger.Infof("[%s] %s (%s)", ui.Plus, ui.Yellow(stage.Id), ui.Grey(stage.Description))
	} else {
		ctx.Logger.Infof("[%s] %s", ui.Plus, ui.Yellow(stage.Id))
	}

}

func RunStage(cfg config.Config, stageCtx *context.Context, stage schema.StageConfig) {

	rootCtx := stageCtx.RootParent()

	var err error
	var cmd *exec.Cmd
	var scriptPath string

	if stage.Script != "" && len(stage.Args) != 0 {
		// both script and args cannot be set simultaneously
		stageCtx.Logger.Fatal("Script and Args cannot be set simultaneously")
	}

	if stage.Script != "" {
		stageCtx.Logger.Debug("Preparing script")
		if stage.Plugin != "" {
			stageCtx.Logger.Fatal("Cannot use both script and plugin")
		}

		tempTargetRunDir := path.Join(stageCtx.TempDir, stage.Id)
		targetRunPath := path.Join(tempTargetRunDir, "run.sh")
		stageCtx.Logger.Debug("Writing script to ", targetRunPath)
		scriptPath = targetRunPath
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

			cwd, err := os.Getwd()
			if err != nil {
				panic(err)
			}
			cmd = exec.Command("podman",
				"run", "--rm", "--entrypoint=sh",
				"-v", fmt.Sprintf("%s:%s:Z", cwd, "/workspace"),
				"-v", tempTargetRunDir+":/workspace.togomak.scripts:Z",
				"-w", "/workspace",
				stage.Container,
				"-c", "/workspace.togomak.scripts/run.sh")
			stageCtx.Logger.Debug("Running ", cmd.String())

		} else {
			cmd = exec.Command("sh", "-c", targetRunPath)
		}
	} else {
		stageCtx.Logger.Tracef("Running with args %v", stage.Args)
		// run the args
		newArgs := make([]string, len(stage.Args))

		// lazy evaluate pongo templates
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

		// no container is specified, no script is specified
		if stage.Container == "" {
			if len(newArgs) == 0 {
				stageCtx.Logger.Fatal("No args specified")
			}

			cmd = exec.Command(newArgs[0], newArgs[1:]...)

		} else
		// container is specified, no script is specified
		{
			stageCtx.Logger.Debug("Preparing container")
			// the user requires a specific container to run on
			cwd, err := os.Getwd()
			if err != nil {
				panic(err)
			}

			dockerArgs := []string{"run", "--rm",
				"-v", fmt.Sprintf("%s:%s:Z", cwd, "/workspace"),
				"-w", "/workspace",
				stage.Container}

			cmd = exec.Command("podman",
				append(dockerArgs, newArgs...)...)
		}

	}

	cmd.Stdout = stageCtx.Logger.Writer()
	cmd.Stderr = stageCtx.Logger.Writer()

	if !cfg.DryRun {
		stageCtx.Logger.Debug("Running ", cmd.String())
		err = cmd.Run()

		// TODO: implement fail fast flag
		if err != nil {
			if !cfg.FailFast {
				stageCtx.Logger.Warnf("Stage failed, continuing because %s.%s=%s", ui.Options, ui.FailFast, ui.False)
				stageCtx.Logger.Warn(err)
			} else {
				stageCtx.Logger.Fatal(err)
			}

		}

	} else {
		if scriptPath != "" {
			fmt.Println(ui.Grey(cmd.String()))
			fmt.Println(ui.Grey("# cat", scriptPath))
			data, err := os.ReadFile(scriptPath)
			if err != nil {
				stageCtx.Logger.Fatal(err)
			}
			fmt.Println(strings.TrimSpace(string(data)))
		} else {
			fmt.Println(cmd.String())
		}
		fmt.Println()
	}
}
