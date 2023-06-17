package ci

import (
	"context"
	"fmt"
	"github.com/alessio/shellescape"
	"github.com/docker/docker/api/types"
	dockerContainer "github.com/docker/docker/api/types/container"
	dockerClient "github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/hashicorp/hcl/v2"
	"github.com/imdario/mergo"
	"github.com/sirupsen/logrus"
	"github.com/srevinsaju/togomak/v1/pkg/c"
	"github.com/srevinsaju/togomak/v1/pkg/diag"
	"github.com/srevinsaju/togomak/v1/pkg/meta"
	"github.com/srevinsaju/togomak/v1/pkg/ui"
	"github.com/zclconf/go-cty/cty"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
)

const TogomakParamEnvVarPrefix = "TOGOMAK__param__"

var TogomakParamEnvVarRegexExpression = fmt.Sprintf("%s([a-zA-Z0-9_]+)", TogomakParamEnvVarPrefix)
var TogomakParamEnvVarRegex = regexp.MustCompile(TogomakParamEnvVarRegexExpression)

func (s *Stage) Prepare(ctx context.Context, skip bool, overridden bool) diag.Diagnostics {
	logger := ctx.Value(c.TogomakContextLogger).(*logrus.Logger)
	// show some user-friendly output on the details of the stage about to be run

	var id string
	if !skip {
		id = ui.Blue(s.Id)
	} else {
		id = fmt.Sprintf("%s %s", ui.Yellow(s.Id), ui.Grey("(skipped)"))
	}
	if overridden {
		id = fmt.Sprintf("%s %s", id, ui.Bold("(overriden)"))
	}
	logger.Infof("[%s] %s", ui.Plus, id)
	return nil
}

func (s *Stage) expandMacros(ctx context.Context) (*Stage, hcl.Diagnostics) {

	if s.Use == nil {
		// this stage does not use a macro
		return s, nil
	}
	hclContext := ctx.Value(c.TogomakContextHclEval).(*hcl.EvalContext)
	logger := ctx.Value(c.TogomakContextLogger).(*logrus.Logger).WithField(StageBlock, s.Id).WithField(MacroBlock, true)
	pipe := ctx.Value(c.TogomakContextPipeline).(*Pipeline)
	cwd := ctx.Value(c.TogomakContextCwd).(string)
	tmpDir := ctx.Value(c.TogomakContextTempDir).(string)
	logger.Debugf("running %s.%s", s.Identifier(), MacroBlock)

	var hclDiags hcl.Diagnostics
	var err error

	v := s.Use.Macro.Variables()
	if v == nil || len(v) == 0 {
		// this stage does not use a macro
		return s, hclDiags
	}

	if len(v) != 1 {
		hclDiags = hclDiags.Append(&hcl.Diagnostic{
			Severity:    hcl.DiagError,
			Summary:     "invalid macro",
			Detail:      fmt.Sprintf("%s can only use a single macro", s.Identifier()),
			EvalContext: hclContext,
			Subject:     v[0].SourceRange().Ptr(),
		})
		return s, hclDiags
	}
	variable := v[0]
	if variable.RootName() != MacroBlock {
		hclDiags = hclDiags.Append(&hcl.Diagnostic{
			Severity:    hcl.DiagError,
			Summary:     "invalid macro",
			Detail:      fmt.Sprintf("%s uses an invalid macro, got '%s'", s.Identifier(), variable.RootName()),
			EvalContext: hclContext,
			Subject:     v[0].SourceRange().Ptr(),
		})
		return s, hclDiags
	}

	macroName := variable[1].(hcl.TraverseAttr).Name
	logger.Debugf("stage.%s uses macro.%s", s.Id, macroName)
	macroRunnable, d := Resolve(ctx, pipe, fmt.Sprintf("macro.%s", macroName))
	if d.HasErrors() {
		d.Fatal(logger.WriterLevel(logrus.ErrorLevel))
	}
	macro := macroRunnable.(*Macro)

	oldStageId := s.Id
	oldStageName := s.Name
	oldStageDependsOn := s.DependsOn

	if macro.Source != "" {
		executable, err := os.Executable()
		if err != nil {
			panic(err)
		}
		parent := shellescape.Quote(s.Id)
		s.Args = hcl.StaticExpr(
			cty.ListVal([]cty.Value{
				cty.StringVal(executable),
				cty.StringVal("--child"),
				cty.StringVal("--dir"), cty.StringVal(cwd),
				cty.StringVal("--file"), cty.StringVal(macro.Source),
				cty.StringVal("--parent"), cty.StringVal(parent),
			}), hcl.Range{Filename: "memory"})

	} else if macro.Stage != nil {
		logger.Debugf("merging %s with %s", s.Identifier(), macro.Identifier())
		err = mergo.Merge(s, macro.Stage, mergo.WithOverride)

	} else {
		f, d := macro.Files.Value(hclContext)
		if d.HasErrors() {
			return s, hclDiags.Extend(d)
		}
		if !f.IsNull() {
			files := f.AsValueMap()
			logger.Debugf("using %d files from %s", len(files), macro.Identifier())
			err = os.MkdirAll(filepath.Join(tmpDir, s.Id), 0755)
			if err != nil {
				return s, hclDiags.Append(&hcl.Diagnostic{
					Severity:    hcl.DiagError,
					Summary:     "failed to create temporary directory",
					Detail:      fmt.Sprintf("failed to create temporary directory for stage %s", s.Id),
					Subject:     variable.SourceRange().Ptr(),
					EvalContext: hclContext,
				})
			}

			defaultExecutionPath := ""
			lastExecutionPath := ""

			for fName, fContent := range files {
				lastExecutionPath = filepath.Join(tmpDir, s.Id, fName)
				if filepath.Base(fName) == meta.ConfigFileName {
					defaultExecutionPath = filepath.Join(tmpDir, s.Id, fName)
				}
				// write the file content to the temporary directory
				// and then add it to the stage
				fpath := filepath.Join(tmpDir, s.Id, fName)
				logger.Debugf("writing %s to %s", fName, fpath)
				err = os.WriteFile(fpath, []byte(fContent.AsString()), 0644)
				if err != nil {
					// TODO: move to diagnostics
					return s, hclDiags.Append(&hcl.Diagnostic{
						Severity:    hcl.DiagError,
						Summary:     "invalid macro",
						Detail:      fmt.Sprintf("%s uses a macro with an invalid file %s", s.Identifier(), fName),
						EvalContext: hclContext,
						Subject:     variable.SourceRange().Ptr(),
					})
				}
			}
			if defaultExecutionPath == "" {
				if len(files) == 1 {
					defaultExecutionPath = lastExecutionPath
				}
			}
			if defaultExecutionPath == "" {
				hclDiags = hclDiags.Append(&hcl.Diagnostic{
					Severity:    hcl.DiagError,
					Summary:     "invalid macro",
					Detail:      fmt.Sprintf("%s uses a macro without a default execution file. include a file named togomak.hcl to avoid this error", s.Identifier()),
					EvalContext: hclContext,
					Subject:     variable.SourceRange().Ptr(),
				})
				return s, hclDiags
			}

			executable, err := os.Executable()
			if err != nil {
				panic(err)
			}
			parent := shellescape.Quote(s.Id)
			s.Args = hcl.StaticExpr(
				cty.ListVal([]cty.Value{
					cty.StringVal(executable),
					cty.StringVal("--child"),
					cty.StringVal("--dir"), cty.StringVal(cwd),
					cty.StringVal("--file"), cty.StringVal(defaultExecutionPath),
					cty.StringVal("--parent"), cty.StringVal(parent),
				}), hcl.Range{Filename: "memory"})

		}
	}

	if err != nil {
		panic(err)
	}
	s.Id = oldStageId
	s.Name = oldStageName
	s.DependsOn = oldStageDependsOn

	return s, nil

}

func (s *Stage) Run(ctx context.Context) diag.Diagnostics {
	hclDgWriter := ctx.Value(c.TogomakContextHclDiagWriter).(hcl.DiagnosticWriter)
	logger := ctx.Value(c.TogomakContextLogger).(*logrus.Logger).WithField(StageBlock, s.Id)
	cwd := ctx.Value(c.TogomakContextCwd).(string)
	logger.Debugf("running %s.%s", StageBlock, s.Id)
	isDryRun := ctx.Value(c.TogomakContextPipelineDryRun).(bool)

	var diags diag.Diagnostics
	var hclDiags hcl.Diagnostics
	var err error
	evalCtx := ctx.Value(c.TogomakContextHclEval).(*hcl.EvalContext)

	// expand stages using macros
	s, d := s.expandMacros(ctx)
	hclDiags = hclDiags.Extend(d)

	paramsGo := map[string]cty.Value{}
	if s.Use != nil && s.Use.Parameters != nil {
		parameters, d := s.Use.Parameters.Value(evalCtx)
		hclDiags = hclDiags.Extend(d)
		if !parameters.IsNull() {
			for k, v := range parameters.AsValueMap() {
				paramsGo[k] = v
			}
		}
	}

	oldParam, ok := evalCtx.Variables["param"]
	if ok {
		oldParamMap := oldParam.AsValueMap()
		for k, v := range oldParamMap {
			paramsGo[k] = v
		}
	}
	evalCtx = evalCtx.NewChild()
	evalCtx.Variables = map[string]cty.Value{
		"this": cty.ObjectVal(map[string]cty.Value{
			"name": cty.StringVal(s.Name),
			"id":   cty.StringVal(s.Id),
		}),
		"param": cty.ObjectVal(paramsGo),
	}

	script, d := s.Script.Value(evalCtx)
	hclDiags = hclDiags.Extend(d)
	args, d := s.Args.Value(evalCtx)
	hclDiags = hclDiags.Extend(d)

	var environment map[string]cty.Value
	environment = make(map[string]cty.Value)
	for _, env := range s.Environment {
		v, d := env.Value.Value(evalCtx)
		hclDiags = hclDiags.Extend(d)
		environment[env.Name] = v
	}

	diags = diags.ExtendHCLDiagnostics(hclDiags, hclDgWriter, s.Identifier())
	if diags.HasErrors() {
		return diags
	}

	envStrings := make([]string, len(environment))
	for k, v := range environment {
		envParsed := fmt.Sprintf("%s=%s", k, v.AsString())
		if isDryRun {
			fmt.Println(ui.Blue("export"), envParsed)
		}

		envStrings = append(envStrings, envParsed)
	}

	if s.Use != nil && s.Use.Parameters != nil {
		for k, v := range paramsGo {
			envParsed := fmt.Sprintf("%s%s=%s", TogomakParamEnvVarPrefix, k, v.AsString())
			if isDryRun {
				fmt.Println(ui.Blue("export"), envParsed)
			}

			envStrings = append(envStrings, envParsed)
		}
	}

	runArgs := make([]string, 0)
	runCommand := "sh"

	// emptyCommands - specifies if both args and scripts were unset
	emptyCommands := false

	if script.Type() == cty.String {
		runArgs = append(runArgs, "-c", script.AsString())
	} else if !args.IsNull() && len(args.AsValueSlice()) != 0 {
		runCommand = args.AsValueSlice()[0].AsString()
		for i, a := range args.AsValueSlice() {
			if i == 0 {
				continue
			}
			runArgs = append(runArgs, a.AsString())
		}
	} else if s.Container == nil {
		fmt.Println(args.Type())
		// if the container is not null, we may rely on internal args or entrypoint scripts
		diags = diags.Append(diag.Diagnostic{
			Severity: diag.SeverityError,
			Summary:  "script or args must be a string",
			Detail:   fmt.Sprintf("the provided script or args, was not recognized as a valid string. received script='''%s''', args='''%s'''", script, args),
		})
		return diags
	} else {
		emptyCommands = true
	}
	dir := cwd
	dirParsed, d := s.Dir.Value(evalCtx)
	if d.HasErrors() {
		hclDiags = hclDiags.Extend(d)
	} else {
		if !dirParsed.IsNull() && dirParsed.AsString() != "" {
			dir = dirParsed.AsString()
		}
		if !filepath.IsAbs(dir) {
			dir = filepath.Join(cwd, dir)
		}
		if isDryRun {
			fmt.Println(ui.Blue("cd"), dir)
		}
	}

	cmd := exec.CommandContext(ctx, runCommand, runArgs...)
	cmd.Stdout = logger.Writer()
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Stderr = logger.WriterLevel(logrus.WarnLevel)
	cmd.Env = append(envStrings, os.Environ()...)
	cmd.Dir = dir

	if s.Container == nil {

		s.process = cmd

		logger.Trace("running command:", cmd.String())

		if !isDryRun {
			err = cmd.Run()
		} else {
			fmt.Println(cmd.String())
		}

	} else {
		logger := logger.WithField("🐳", "")
		cli, err := dockerClient.NewClientWithOpts(dockerClient.FromEnv, dockerClient.WithAPIVersionNegotiation())
		if err != nil {
			return diags.Append(diag.Diagnostic{
				Severity: diag.SeverityError,
				Summary:  "could not create docker client",
				Detail:   err.Error(),
				Source:   s.Identifier(),
			})
		}
		defer cli.Close()
		// check if image exists
		logger.Debugf("checking if image %s exists", s.Container.Image)
		_, _, err = cli.ImageInspectWithRaw(ctx, s.Container.Image)
		if err != nil {
			logger.Infof("image %s does not exist, pulling...", s.Container.Image)
			reader, err := cli.ImagePull(ctx, s.Container.Image, types.ImagePullOptions{})
			if err != nil {
				return diags.Append(diag.Diagnostic{
					Severity: diag.SeverityError,
					Summary:  "could not pull image",
					Detail:   err.Error(),
					Source:   s.Identifier(),
				})
			}

			pb := ui.NewProgressWriter(reader, logger.Writer(), fmt.Sprintf("pulling image %s", s.Container.Image))
			defer pb.Close()
			defer reader.Close()
			io.Copy(pb, reader)

		}

		var containerArgs []string
		if !emptyCommands {
			containerArgs = cmd.Args
		}
		binds := []string{
			fmt.Sprintf("%s:/workspace", cmd.Dir),
		}

		for _, m := range s.Container.Volumes {
			source, d := m.Source.Value(evalCtx)
			hclDiags = hclDiags.Extend(d)
			dest, d := m.Destination.Value(evalCtx)
			hclDiags = hclDiags.Extend(d)
			if hclDiags.HasErrors() {
				continue
			}
			binds = append(binds, fmt.Sprintf("%s:%s", source.AsString(), dest.AsString()))
		}
		diags = diags.ExtendHCLDiagnostics(hclDiags, hclDgWriter, s.Identifier())
		if diags.HasErrors() {
			return diags
		}

		if !isDryRun {
			resp, err := cli.ContainerCreate(ctx, &dockerContainer.Config{
				Image:        s.Container.Image,
				Cmd:          containerArgs,
				WorkingDir:   "/workspace",
				Tty:          true,
				AttachStdout: true,
				AttachStderr: true,
				AttachStdin:  s.Container.Stdin,
				OpenStdin:    s.Container.Stdin,
				StdinOnce:    s.Container.Stdin,
				Env:          envStrings,
				// User: s.Container.User,
			}, &dockerContainer.HostConfig{
				Binds: binds,
			}, nil, nil, "")
			if err != nil {
				return diags.Append(diag.Diagnostic{
					Severity: diag.SeverityError,
					Summary:  "could not create container",
					Detail:   err.Error(),
					Source:   s.Identifier(),
				})
			}

			if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
				return diags.Append(diag.Diagnostic{
					Severity: diag.SeverityError,
					Summary:  "could not start container",
					Detail:   err.Error(),
					Source:   s.Identifier(),
				})
			}
			s.ContainerId = resp.ID

			container, err := cli.ContainerInspect(ctx, resp.ID)
			if err != nil {
				panic(err)
			}

			responseBody, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{
				ShowStdout: true, ShowStderr: true,
				Follow: true,
			})
			if err != nil {
				diags = diags.Append(diag.Diagnostic{
					Severity: diag.SeverityError,
					Summary:  "failed to get container logs",
					Detail:   err.Error(),
				})
				diags.Write(logger.WriterLevel(logrus.ErrorLevel))
				return diags
			}
			defer responseBody.Close()

			if container.Config.Tty {
				_, err = io.Copy(logger.Writer(), responseBody)
			} else {
				_, err = stdcopy.StdCopy(logger.Writer(), logger.WriterLevel(logrus.WarnLevel), responseBody)
			}
			if err != nil && err != io.EOF {
				if err == context.Canceled {
					return diags
				}
				diags = diags.Append(diag.Diagnostic{
					Severity: diag.SeverityError,
					Summary:  "failed to copy container logs",
					Detail:   err.Error(),
				})
			}
			err = cli.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{
				RemoveVolumes: true,
			})
			if err != nil {
				diags = diags.Append(diag.Diagnostic{
					Severity: diag.SeverityWarning,
					Summary:  "failed to remove docker container",
					Detail:   err.Error(),
				})
				return diags
			}
		} else {
			fmt.Println(ui.Blue("docker:run.image"), ui.Green(s.Container.Image))

			fmt.Println(ui.Blue("docker:run.workdir"), ui.Green("/workspace"))
			fmt.Println(ui.Blue("docker:run.volume"), ui.Green(cmd.Dir+":/workspace"))

			fmt.Println(ui.Blue("docker:run.env"), ui.Green(strings.Join(envStrings, " ")))
			fmt.Println(ui.Blue("docker:run.stdin"), ui.Green(s.Container.Stdin))
			fmt.Println(ui.Blue("docker:run.args"), ui.Green(strings.Join(containerArgs, " ")))

		}

	}

	if err != nil {
		diags = diags.Append(diag.Diagnostic{
			Severity: diag.SeverityError,
			Summary:  "failed to run command",
			Detail:   err.Error(),
			Source:   s.Identifier(),
		})
		return diags
	}

	return diags
}

func (s *Stage) CanRun(ctx context.Context) (bool, diag.Diagnostics) {
	logger := ctx.Value(c.TogomakContextLogger).(*logrus.Logger).WithField("stage", s.Id)
	logger.Debugf("checking if stage.%s can run", s.Id)

	var diags diag.Diagnostics
	evalCtx := ctx.Value(c.TogomakContextHclEval).(*hcl.EvalContext)
	hclWriter := ctx.Value(c.TogomakContextHclDiagWriter).(hcl.DiagnosticWriter)
	var hclDiags hcl.Diagnostics

	paramsGo := map[string]cty.Value{}
	if s.Use != nil && s.Use.Parameters != nil {
		parameters, d := s.Use.Parameters.Value(evalCtx)
		hclDiags = hclDiags.Extend(d)
		if !parameters.IsNull() {
			for k, v := range parameters.AsValueMap() {
				paramsGo[k] = v
			}
		}
	}

	evalCtx = evalCtx.NewChild()
	evalCtx.Variables = map[string]cty.Value{
		"this": cty.ObjectVal(map[string]cty.Value{
			"name": cty.StringVal(s.Name),
			"id":   cty.StringVal(s.Id),
		}),
		"param": cty.ObjectVal(paramsGo),
	}
	v, hclDiags := s.Condition.Value(evalCtx)
	if hclDiags.HasErrors() {
		err := hclWriter.WriteDiagnostics(hclDiags)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.SeverityError,
				Summary:  "failed to write HCL diagnostics",
				Detail:   err.Error(),
			})
		}

		diags = diags.Append(diag.Diagnostic{
			Severity: diag.SeverityError,
			Summary:  "failed to evaluate condition",
			Detail:   hclDiags.Error(),
			Source:   "stage_condition_check",
		})

		return false, diags
	}
	if v.Equals(cty.False).True() {
		// this stage has been explicitly evaluated to false
		// we will not run this
		return false, diags
	}

	return true, diags
}

func dockerContainerSourceFmt(containerId string) string {
	return fmt.Sprintf("docker: container=%s", containerId)
}

func (s *Stage) Terminate() diag.Diagnostics {
	var diags diag.Diagnostics
	if s.Container != nil && s.ContainerId != "" {
		ctx := context.Background()

		cli, err := dockerClient.NewClientWithOpts(dockerClient.FromEnv, dockerClient.WithAPIVersionNegotiation())
		if err != nil {
			diags = diags.Append(diag.Diagnostic{
				Severity: diag.SeverityError,
				Summary:  "failed to create docker client",
				Detail:   err.Error(),
				Source:   dockerContainerSourceFmt(s.ContainerId),
			})
		}
		err = cli.ContainerStop(ctx, s.ContainerId, dockerContainer.StopOptions{})
		if err != nil {
			diags = diags.Append(diag.Diagnostic{
				Severity: diag.SeverityError,
				Summary:  "failed to stop container",
				Detail:   err.Error(),
				Source:   dockerContainerSourceFmt(s.ContainerId),
			})
		}
	} else if s.process != nil {
		err := s.process.Process.Signal(syscall.SIGTERM)
		if err != nil {
			diags = diags.Append(diag.Diagnostic{
				Severity: diag.SeverityError,
				Summary:  "failed to terminate process",
				Source:   s.Identifier(),
				Detail:   err.Error(),
			})
		}
	}

	return diags
}

func (s *Stage) Kill() diag.Diagnostics {
	diags := s.Terminate()
	if s.process != nil && !s.process.ProcessState.Exited() {
		err := s.process.Process.Kill()
		if err != nil {
			diags = diags.Append(diag.Diagnostic{
				Severity: diag.SeverityError,
				Source:   s.Identifier(),
				Summary:  "couldn't kill stage",
				Detail:   err.Error(),
			})
		}
	}
	return diags
}
