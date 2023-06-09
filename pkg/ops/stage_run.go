package ops

import (
	"errors"
	"fmt"
	"github.com/bcicen/jstream"
	"github.com/flosch/pongo2/v6"
	"github.com/hashicorp/go-envparse"
	"github.com/spf13/afero"
	"github.com/srevinsaju/togomak/pkg/config"
	"github.com/srevinsaju/togomak/pkg/meta"
	"github.com/srevinsaju/togomak/pkg/sources"
	"github.com/srevinsaju/togomak/pkg/templating"
	"github.com/srevinsaju/togomak/pkg/ui"
	"github.com/srevinsaju/togomak/pkg/x"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/srevinsaju/togomak/pkg/context"
	"github.com/srevinsaju/togomak/pkg/schema"
)

var rightArrow = ui.Grey("→")
var stageConstText = ui.HiCyan("stage")

func childTogomakReader(reader io.Reader, stageCtx *context.Context) {
	decoder := jstream.NewDecoder(reader, 0)
	for mv := range decoder.Stream() {

		msg := strings.Builder{}
		mv, ok := mv.Value.(map[string]interface{})
		if !ok {
			continue
		}
		if v, ok := mv["stage"].(string); ok && v != "" {
			msg.WriteString(fmt.Sprintf("%s [%s=%s] ", rightArrow, stageConstText, mv["stage"].(string)))
		}
		if v, ok := mv["msg"].(string); ok && v != "" {
			msg.WriteString(mv["msg"].(string))
		}
		switch mv["level"].(string) {
		case "info":
			stageCtx.Logger.Info(msg.String())
			break
		case "error":
			stageCtx.Logger.Error(msg.String())
			break
		case "warn":
			stageCtx.Logger.Warn(msg.String())
			break
		case "debug":
			stageCtx.Logger.Debug(msg.String())
			break
		case "trace":
			stageCtx.Logger.Trace(msg.String())
			break
		default:
			stageCtx.Logger.Info(msg.String())
			break
		}
	}

}

func PrepareStage(ctx *context.Context, stage *schema.StageConfig, skipped bool) {
	// show some user-friendly output on the details of the stage about to be run
	var name string
	var id string
	if !skipped {
		name = ui.Blue(stage.Name)
		id = ui.Blue(stage.Id)
	} else {
		name = ui.Yellow(stage.Id)
		id = ui.Yellow(stage.Id)
	}

	if stage.Name != "" && stage.Description != "" {
		ctx.Logger.Infof("[%s] %s (%s)", ui.Plus, name, ui.Grey(stage.Description))
	} else if stage.Name != "" {
		ctx.Logger.Infof("[%s] %s", ui.Plus, name)
	} else if stage.Description != "" {
		ctx.Logger.Infof("[%s] %s (%s)", ui.Plus, id, ui.Grey(stage.Description))
	} else {
		ctx.Logger.Infof("[%s] %s", ui.Plus, id)
	}

}

func RunStage(cfg config.Config, stageCtx *context.Context, stage schema.StageConfig) error {

	rootCtx := stageCtx.RootParent()

	var err error
	var cmd *exec.Cmd

	var scriptPath string

	if stage.Script != "" && len(stage.Args) != 0 {
		// both script and args cannot be set simultaneously
		return errors.New(".script and .args cannot be set simultaneously")
	}

	if stage.Script != "" {
		stageCtx.Logger.Debug("preparing script")
		if stage.Plugin != "" {
			return errors.New("cannot use both script and plugin")
		}

		tempTargetRunDir := path.Join(stageCtx.TempDir, stage.Id)
		targetRunPath := path.Join(tempTargetRunDir, "run.sh")
		stageCtx.Logger.Debug("Writing script to ", targetRunPath)
		scriptPath = targetRunPath
		err = os.MkdirAll(tempTargetRunDir, 0755)
		if err != nil {
			return err
		}

		tpl, err := pongo2.FromString(stage.Script)
		if err != nil {
			return err
		}
		data, err := templating.ExecuteWithStage(tpl, stageCtx.RootParent().Data, stage)
		if err != nil {
			return err
		}

		err = os.WriteFile(targetRunPath, []byte(data), 0755)
		if err != nil {
			return err
		}

		if stage.Container != "" {
			stageCtx.Logger.Debug("Preparing container")
			// the user requires a specific container to run on

			cwd, err := os.Getwd()
			if err != nil {
				panic(err)
			}
			cmd = exec.Command("docker",
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
	} else if stage.Source.Type != "" {
		stageCtx.Logger.Debug("Preparing source")
		togomakContextDir := cfg.ContextDir
		if togomakContextDir == "" {
			togomakContextDir = "."
		}
		togomakYamlDir := sources.GetStorePath(rootCtx, stage)
		togomakRelativeYamlFilePath := stage.Source.File
		if togomakRelativeYamlFilePath == "" {
			togomakRelativeYamlFilePath = "togomak.yaml"
		}
		togomakYamlPath := filepath.Join(togomakYamlDir, togomakRelativeYamlFilePath)
		togomak := x.MustReturn(os.Executable()).(string)

		args := []string{
			"--context", togomakContextDir,
			"--file", togomakYamlPath,
			"--jobs", strconv.Itoa(cfg.JobsNumber),
			"--summary=false",
			"--child",
			"--no-interactive",
			// TODO: add --ci flag
			// TODO: add color flag
			// TODO: add --debug flag
			fmt.Sprintf("--fail-lazy=%v", cfg.FailLazy),
		}
		if cfg.DryRun {
			args = append(args, "--dry-run")
		}
		for _, param := range stage.Source.Parameters {
			v := []string{
				"--parameters",
				fmt.Sprintf("%s=%s", param.Name, param.Default),
			}
			args = append(args, v...)
		}

		if stage.Source.Stages != nil {
			args = append(args, stage.Source.Stages...)
		}

		cmd = exec.Command(togomak, args...)
		stageCtx.Logger.Warn("Running ", cmd.String())

	} else {
		stageCtx.Logger.Tracef("Running with args %v", stage.Args)
		// run the args
		newArgs := make([]string, len(stage.Args))

		// lazy evaluate pongo templates
		for i, arg := range stage.Args {
			// render them with pongo
			tpl, err := pongo2.FromString(arg)
			if err != nil {
				return fmt.Errorf("cannot render args '%s': %v", arg, err)
			}
			parsed, err := templating.ExecuteWithStage(tpl, rootCtx.Data.AsMap(), stage)
			if err != nil {
				return fmt.Errorf("cannot render args '%s': %v", arg, err)
			}
			newArgs[i] = parsed

		}

		// no container is specified, no script is specified
		if stage.Container == "" {
			if len(newArgs) == 0 {
				return fmt.Errorf("no args specified")
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

			cmd = exec.Command("docker",
				append(dockerArgs, newArgs...)...)
		}

	}

	rootCtx.AddProcess(&context.RunningStage{
		Id:      stage.Id,
		Process: cmd,
	})

	cmd.Stdout = stageCtx.Logger.Writer()
	cmd.Stderr = stageCtx.Logger.Writer()
	cmd.Env = os.Environ()
	if stage.Dir != "" {
		tpl, err := pongo2.FromString(stage.Dir)
		if err != nil {
			return fmt.Errorf("cannot render dir '%s': %v", stage.Dir, err)
		}
		parsed, err := templating.ExecuteWithStage(tpl, rootCtx.Data.AsMap(), stage)
		if filepath.IsAbs(parsed) {
			cmd.Dir = parsed
		} else {
			cwd, err := os.Getwd()
			if err != nil {
				panic(err)
			}
			cmd.Dir = filepath.Join(cwd, parsed)
		}
		stageCtx.Logger.Debugf("command will be executed in %s", cmd.Dir)
	}

	// add output storage
	f, err := afero.TempFile(afero.NewOsFs(), rootCtx.TempDir, stage.Id)
	if err != nil {
		return fmt.Errorf("cannot create temp file: %v", err)
	}
	defer f.Close()
	cmd.Env = append(cmd.Env, fmt.Sprintf("TOGOMAK_ENV=%s", f.Name()))

	// add togomak build Id
	cmd.Env = append(cmd.Env, fmt.Sprintf("TOGOMAK_BUILD_ID=%s", meta.GetCorrelationId()))

	// add root temp dir location
	cmd.Env = append(cmd.Env, fmt.Sprintf("TOGOMAK_TEMPDIR=%s", rootCtx.TempDir))

	// add environment variables
	for k, v := range stage.Environment {
		tpl, err := pongo2.FromString(v)
		if err != nil {
			return fmt.Errorf("cannot render args '%s': %v", v, err)
		}
		parsedV, err := templating.ExecuteWithStage(tpl, rootCtx.Data.AsMap(), stage)
		parsedV = strings.TrimSpace(parsedV)

		if err != nil {
			return fmt.Errorf("cannot render args '%s': %v", v, err)
		}

		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, parsedV))
		if cfg.DryRun {
			fmt.Printf("export %s=%s\n", k, parsedV)
		}
	}

	// we will be reading input from togomak, so we need rich output
	if stage.Source.Type != "" {

		pr, pw := io.Pipe()

		go childTogomakReader(pr, stageCtx)

		cmd.Stdout = pw
		cmd.Stderr = os.Stderr
	}
	matrixId := ""
	if rootCtx.IsMatrix {
		m := rootCtx.Data["matrix"].(map[string]string)
		for k, v := range m {
			// TODO: check if this will create any parsing / escape issues
			matrixId += fmt.Sprintf("%s=%v, ", k, v)
		}
	}
	stageCtx.Logger.Debugf("Detected matrix stage %s", matrixId)
	status := &context.Status{
		MatrixId: matrixId,
	}

	if !cfg.DryRun {
		stageCtx.Logger.Debug("Running ", cmd.String())

		var err error
		stageCtx.Logger.Trace("running stage with retry", stage.Retry.Enabled)
		if stage.Retry.Enabled {

			operation := func() error {

				err := cmd.Run()
				if err != nil {
					stageCtx.Logger.Info("Stage failed, retrying...")
					return err
				}
				return nil
			}
			err = operation()
			if err != nil {
				fmt.Println(err)
				err = operation()
				if err != nil {
					fmt.Println(err)
				}
			}

			//expBackOff := backoff.NewExponentialBackOff()
			//			err = backoff.Retry(operation, expBackOff)
		} else {
			err = cmd.Run()

		}

		// TODO: implement fail fast flag
		if err != nil {
			status.Message = err.Error()
			if cfg.FailLazy {
				stageCtx.Logger.Warnf("Stage failed, continuing because %s.%s=%s", ui.Options, ui.FailLazy, ui.False)
				stageCtx.Logger.Warn(err)
			} else {
				stageCtx.Logger.Warn(err)
				return err
			}

		} else {
			status.Success = true
		}

	} else {
		if scriptPath != "" {
			fmt.Println(ui.Grey(cmd.String()))
			fmt.Println(ui.Grey("# cat ", scriptPath))
			data, err := os.ReadFile(scriptPath)
			if err != nil {
				return err
			}
			fmt.Println(strings.TrimSpace(string(data)))
		} else {
			fmt.Println(cmd.String())
		}
		fmt.Println()
		status.Success = true
	}

	// parse saved environment file
	s, err := envparse.Parse(f)
	if err != nil {
		stageCtx.Logger.Warn(err)
		return err
	}
	stageCtx.DataMutex.Lock()
	stageCtx.Data["env"] = s
	stageCtx.DataMutex.Unlock()

	stageCtx.SetStatus(*status)
	return nil
}
