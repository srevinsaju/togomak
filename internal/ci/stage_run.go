package ci

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/alessio/shellescape"
	"github.com/docker/docker/api/types"
	dockerContainer "github.com/docker/docker/api/types/container"
	dockerClient "github.com/docker/docker/client"
	"github.com/srevinsaju/togomak/v1/internal/blocks"
	"github.com/srevinsaju/togomak/v1/internal/dg"
	"github.com/srevinsaju/togomak/v1/internal/x"
	"sync"

	"github.com/docker/docker/pkg/stdcopy"
	"github.com/google/uuid"
	"github.com/hashicorp/hcl/v2"
	"github.com/imdario/mergo"
	"github.com/sirupsen/logrus"
	"github.com/srevinsaju/togomak/v1/internal/c"
	"github.com/srevinsaju/togomak/v1/internal/meta"
	"github.com/srevinsaju/togomak/v1/internal/runnable"
	"github.com/srevinsaju/togomak/v1/internal/ui"
	"github.com/zclconf/go-cty/cty"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

const TogomakParamEnvVarPrefix = "TOGOMAK__param__"

var TogomakParamEnvVarRegexExpression = fmt.Sprintf("%s([a-zA-Z0-9_]+)", TogomakParamEnvVarPrefix)
var TogomakParamEnvVarRegex = regexp.MustCompile(TogomakParamEnvVarRegexExpression)

func (s *Stage) Prepare(conductor *Conductor, skip bool, overridden bool) hcl.Diagnostics {
	logger := conductor.Logger().WithField("stage", s.Id)
	// show some user-friendly output on the details of the stage about to be run

	var id string
	if !skip {
		id = ""
	} else {
		id = fmt.Sprintf("%s", ui.Grey("skipped"))
	}
	if overridden {
		id = fmt.Sprintf("%s", ui.Blue("overridden"))
	}
	logger.Infof("%s", id)
	return nil
}

// expandMacros expands the macro in the stage, if any.
func (s *Stage) expandMacros(conductor *Conductor, opts ...runnable.Option) (*Stage, hcl.Diagnostics) {
	logger := conductor.Logger().WithField("stage", s.Id)
	cfg := runnable.NewConfig(opts...)
	ctx := conductor.Context()

	if s.Use == nil {
		// this stage does not use a macro
		return s, nil
	}
	hclContext := conductor.Eval().Context()

	pipe := ctx.Value(c.TogomakContextPipeline).(*Pipeline)

	tmpDir := conductor.TempDir()

	logger.Debugf("running %s.%s", s.Identifier(), blocks.MacroBlock)

	var diags hcl.Diagnostics
	var err error

	v := s.Use.Macro.Variables()

	var macro *Macro
	if len(v) == 1 && v[0].RootName() == blocks.MacroBlock {
		variable := v[0]
		macroName := variable[1].(hcl.TraverseAttr).Name
		logger.Debugf("stage.%s uses macro.%s", s.Id, macroName)
		macroRunnable, d := Resolve(pipe, fmt.Sprintf("macro.%s", macroName))
		if d.HasErrors() {
			return nil, diags.Extend(d)
		}
		macro = macroRunnable.(*Macro)
	} else {
		conductor.Eval().Mutex().RLock()
		source, d := s.Use.Macro.Value(hclContext)
		conductor.Eval().Mutex().RUnlock()

		if d.HasErrors() {
			return s, diags.Extend(d)
		}
		if source.IsNull() {
			return s, diags.Append(&hcl.Diagnostic{
				Severity:    hcl.DiagError,
				Summary:     "invalid macro",
				Detail:      fmt.Sprintf("%s uses a macro with an invalid source", s.Identifier()),
				EvalContext: hclContext,
				Subject:     s.Use.Macro.Range().Ptr(),
			})
		}
		if source.Type() != cty.String {
			return s, diags.Append(&hcl.Diagnostic{
				Severity:    hcl.DiagError,
				Summary:     "invalid macro",
				Detail:      fmt.Sprintf("%s uses a macro with an invalid source", s.Identifier()),
				EvalContext: hclContext,
				Subject:     s.Use.Macro.Range().Ptr(),
			})
		}
		sourceEvaluated := source.AsString()
		macro = &Macro{
			Source: sourceEvaluated,
			Id:     uuid.New().String(),
		}
	}

	oldStageId := s.Id
	oldStageName := s.Name
	oldStageDependsOn := s.DependsOn

	// chdir
	conductor.Eval().Mutex().RLock()
	chdirRaw, d := s.Use.Chdir.Value(hclContext)
	conductor.Eval().Mutex().RUnlock()
	if d.HasErrors() {
		return s, diags.Extend(d)
	}
	chdir := true
	if chdirRaw.IsNull() {
		chdir = false
	} else if chdirRaw.Type() == cty.Bool {
		chdir = chdirRaw.True()
	}
	dir := cfg.Paths.Cwd
	if chdir {
		dir = filepath.Join(tmpDir, s.Id)
	}

	if macro.Source != "" {
		executable, err := os.Executable()
		if err != nil {
			panic(err)
		}
		parent := shellescape.Quote(s.Id)

		conductor.Eval().Mutex().RLock()
		stageDir, d := s.Dir.Value(hclContext)
		conductor.Eval().Mutex().RUnlock()
		if d.HasErrors() {
			return s, diags.Extend(d)
		}
		if stageDir.Type() != cty.String && !stageDir.IsNull() {
			return s, diags.Append(&hcl.Diagnostic{
				Severity:    hcl.DiagError,
				Summary:     "invalid stage",
				Detail:      fmt.Sprintf("%s uses a stage with an invalid dir", s.Identifier()),
				EvalContext: hclContext,
				Subject:     s.Dir.Range().Ptr(),
			})
		}

		if stageDir.IsNull() || stageDir.AsString() == "" {
			s.Dir = hcl.StaticExpr(cty.StringVal(cfg.Paths.Cwd), hcl.Range{Filename: "memory"})
		}

		src := macro.Source

		if strings.HasSuffix(macro.Source, ".hcl") {
			diags = diags.Append(&hcl.Diagnostic{
				Severity:    hcl.DiagWarning,
				Summary:     "deprecated",
				Detail:      fmt.Sprintf("%s uses a macro with a .hcl file. use a directory instead", s.Identifier()),
				EvalContext: hclContext,
				Subject:     s.Use.Macro.Range().Ptr(),
			})
			logger.Warnf("macro.Source pointing to a .hcl file is deprecated since v1.6.0. use a directory instead")
			src = filepath.Dir(macro.Source)

			srcAbs, srcErr := filepath.Abs(src)
			cwdAbs, cwdErr := filepath.Abs(cfg.Paths.Cwd)
			if srcErr != nil {
				panic(srcErr)
			}
			if cwdErr != nil {
				panic(cwdErr)
			}
			if srcAbs == cwdAbs {
				return nil, diags.Append(&hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "infinite recursion",
					Detail:   fmt.Sprintf("%s uses a macro with a source pointing to the same directory as the current directory", s.Identifier()),
					Context:  s.Use.Macro.Range().Ptr(),
				})
			}

		}

		s.Args = hcl.StaticExpr(
			cty.ListVal([]cty.Value{
				cty.StringVal(executable),
				cty.StringVal("--child"),
				cty.StringVal("--dir"), cty.StringVal(src),
				cty.StringVal("--parent"), cty.StringVal(parent),
			}), hcl.Range{Filename: "memory"})

	} else if macro.Stage != nil {
		logger.Debugf("merging %s with %s", s.Identifier(), macro.Identifier())
		err = mergo.Merge(s, macro.Stage, mergo.WithOverride)
		s.dependsOnVariablesMacro = macro.Stage.DependsOn.Variables()

	} else {
		conductor.Eval().Mutex().RLock()
		f, d := macro.Files.Value(hclContext)
		conductor.Eval().Mutex().RUnlock()
		if d.HasErrors() {
			return s, diags.Extend(d)
		}

		if !f.IsNull() {
			files := f.AsValueMap()
			logger.Debugf("using %d files from %s", len(files), macro.Identifier())
			err = os.MkdirAll(filepath.Join(tmpDir, s.Id), 0755)
			if err != nil {
				return s, diags.Append(&hcl.Diagnostic{
					Severity:    hcl.DiagError,
					Summary:     "failed to create temporary directory",
					Detail:      fmt.Sprintf("failed to create temporary directory for stage %s", s.Id),
					Subject:     s.Use.Macro.Range().Ptr(),
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
				if fContent.IsNull() {
					return s, diags.Append(&hcl.Diagnostic{
						Severity:    hcl.DiagError,
						Summary:     "invalid macro",
						Detail:      fmt.Sprintf("%s uses a macro with an invalid file %s", s.Identifier(), fName),
						EvalContext: hclContext,
						Subject:     s.Use.Macro.Range().Ptr(),
					})
				}
				err = os.WriteFile(fpath, []byte(fContent.AsString()), 0644)
				if err != nil {
					// TODO: move to diagnostics
					return s, diags.Append(&hcl.Diagnostic{
						Severity:    hcl.DiagError,
						Summary:     "invalid macro",
						Detail:      fmt.Sprintf("%s uses a macro with an invalid file %s", s.Identifier(), fName),
						EvalContext: hclContext,
						Subject:     s.Use.Macro.Range().Ptr(),
					})
				}
			}
			if defaultExecutionPath == "" {
				if len(files) == 1 {
					defaultExecutionPath = lastExecutionPath
				}
			}
			if defaultExecutionPath == "" {
				diags = diags.Append(&hcl.Diagnostic{
					Severity:    hcl.DiagError,
					Summary:     "invalid macro",
					Detail:      fmt.Sprintf("%s uses a macro without a default execution file. include a file named togomak.hcl to avoid this error", s.Identifier()),
					EvalContext: hclContext,
					Subject:     s.Use.Macro.Range().Ptr(),
				})
				return s, diags
			}

			executable, err := os.Executable()
			if err != nil {
				panic(err)
			}
			parent := shellescape.Quote(s.Id)
			args := []cty.Value{
				cty.StringVal(executable),
				cty.StringVal("--child"),
				cty.StringVal("--dir"), cty.StringVal(dir),
				cty.StringVal("--file"), cty.StringVal(defaultExecutionPath),
				cty.StringVal("--parent"), cty.StringVal(parent),
			}
			if cfg.Behavior.Ci {
				args = append(args, cty.StringVal("--ci"))
			}
			if cfg.Behavior.Unattended {
				args = append(args, cty.StringVal("--unattended"))
			}
			childStatuses := s.Get(StageContextChildStatuses).([]string)
			logger.Trace("child statuses: ", childStatuses)
			if childStatuses != nil {
				var ctyChildStatuses []cty.Value
				for _, childStatus := range childStatuses {
					ctyChildStatuses = append(ctyChildStatuses, cty.StringVal(childStatus))
				}
				args = append(args, ctyChildStatuses...)
			}
			s.Args = hcl.StaticExpr(
				cty.ListVal(args),
				hcl.Range{Filename: "memory"})

		}
	}

	if err != nil {
		panic(err)
	}
	s.Id = oldStageId
	s.Name = oldStageName
	s.DependsOn = oldStageDependsOn
	s.dependsOnVariablesMacro = append(s.dependsOnVariablesMacro, oldStageDependsOn.Variables()...)

	return s, nil

}

func (s *Stage) Run(conductor *Conductor, options ...runnable.Option) (diags hcl.Diagnostics) {
	logger := conductor.Logger().WithField("stage", s.Id)
	cfg := runnable.NewConfig(options...)

	logger.Debugf("running %s", x.RenderBlock(blocks.StageBlock, s.Id))

	evalCtx := conductor.Eval().Context()

	// expand stages using macros
	logger.Debugf("expanding macros")
	s, d := s.expandMacros(conductor, options...)
	diags = diags.Extend(d)
	logger.Debugf("finished expanding macros with %d errors", len(diags.Errs()))

	if s.ForEach == nil {
		d = s.run(conductor, evalCtx, options...)
		diags = diags.Extend(d)
		return diags
	}

	conductor.Eval().Mutex().RLock()
	forEachItems, d := s.ForEach.Value(evalCtx)
	conductor.Eval().Mutex().RUnlock()

	diags = diags.Extend(d)
	if d.HasErrors() {
		return diags
	}

	if forEachItems.IsNull() {
		d = s.run(conductor, evalCtx, options...)
		diags = diags.Extend(d)
		return diags
	}
	if !forEachItems.CanIterateElements() {
		diags = diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "invalid type for for_each",
			Detail:   fmt.Sprintf("for_each must be a set or map of objects"),
		})
		return diags
	}

	var wg sync.WaitGroup

	var safeDg dg.SafeDiagnostics

	var counter int
	var key string
	var keyCty cty.Value
	forEachItems.ForEachElement(func(k cty.Value, v cty.Value) bool {
		if k.Type() == cty.String {
			key = fmt.Sprintf("\"%s\"", k.AsString())
			keyCty = k
		} else {
			key = fmt.Sprintf("%d", counter)
			keyCty = cty.NumberIntVal(int64(counter))
		}
		counter++
		id := fmt.Sprintf("%s[%s]", s.Id, key)
		wg.Add(1)
		stage := &Stage{Id: id, CoreStage: s.CoreStage, Lifecycle: s.Lifecycle}
		go func(keyCty cty.Value, options ...runnable.Option) {
			options = append(options, runnable.WithEach(keyCty, v))
			d := stage.Run(conductor, options...)
			safeDg.Extend(d)
			wg.Done()
		}(keyCty, options...)
		if cfg.Behavior.DisableConcurrency {
			wg.Wait()
		}
		return false
	})
	wg.Wait()
	return safeDg.Diagnostics()

}

func (s *Stage) run(conductor *Conductor, evalCtx *hcl.EvalContext, options ...runnable.Option) hcl.Diagnostics {
	var err error
	logger := conductor.Logger().WithField("stage", s.Id)
	tmpDir := conductor.TempDir()
	status := runnable.StatusRunning
	cfg := runnable.NewConfig(options...)
	stream := conductor.NewOutputMemoryStream(s.String())
	diags := &dg.Diagnostics{}

	defer func(stream *bytes.Buffer) {
		logger.Debug("running post hooks")
		success := !diags.HasErrors()
		if !success {
			status = runnable.StatusFailure
		} else {
			status = runnable.StatusSuccess
		}
		hookOpts := []runnable.Option{
			runnable.WithStatus(status),
			runnable.WithHook(),
			runnable.WithStatusOutput(stream.String()),
			runnable.WithParent(runnable.ParentConfig{Name: s.Name, Id: s.Id}),
		}
		hookOpts = append(hookOpts, options...)
		diags.Extend(s.AfterRun(conductor, hookOpts...))
		logger.Debug("finished running post hooks")
	}(stream)

	d := s.executePreHooks(conductor, status, options...)
	diags.Extend(d)

	paramsGo := map[string]cty.Value{}

	logger.Debugf("expanding global macro parameters")
	conductor.Eval().Mutex().RLock()
	oldParam, ok := evalCtx.Variables[blocks.ParamBlock]
	conductor.Eval().Mutex().RUnlock()
	if ok {
		oldParamMap := oldParam.AsValueMap()
		for k, v := range oldParamMap {
			paramsGo[k] = v
		}
	}

	id := s.Id
	name := s.Name
	if cfg.Parent != nil {
		logger.Debugf("using parent %s.%s", cfg.Parent.Name, cfg.Parent.Id)
		id = cfg.Parent.Id
		name = cfg.Parent.Name
	}

	logger.Debug("creating new evaluation context")
	evalCtx = evalCtx.NewChild()
	evalCtx.Variables = map[string]cty.Value{
		ThisBlock: cty.ObjectVal(map[string]cty.Value{
			"name":   cty.StringVal(name),
			"id":     cty.StringVal(id),
			"hook":   cty.BoolVal(cfg.Hook),
			"status": cty.StringVal(string(cfg.Status.Status)),
			"output": cty.StringVal(cfg.Status.Output),
		}),
	}
	if cfg.Each != nil {
		evalCtx.Variables[EachBlock] = cty.ObjectVal(cfg.Each)
	}

	logger.Debugf("expanding macro parameters")
	if s.Use != nil && s.Use.Parameters != nil {
		conductor.Eval().Mutex().RLock()
		parameters, d := s.Use.Parameters.Value(evalCtx)
		conductor.Eval().Mutex().RUnlock()
		diags.Extend(d)
		if !parameters.IsNull() {
			for k, v := range parameters.AsValueMap() {
				paramsGo[k] = v
			}
		}
	}
	evalCtx.Variables[blocks.ParamBlock] = cty.ObjectVal(paramsGo)

	environment, d := s.parseEnvironmentVariables(conductor, evalCtx)
	diags.Extend(d)
	if diags.HasErrors() {
		return diags.Diagnostics()
	}

	envStrings := s.processEnvironmentVariables(conductor, environment, cfg, tmpDir, paramsGo)

	cmd, d := s.parseExecCommand(conductor, evalCtx, cfg, envStrings, stream)
	diags.Extend(d)
	if diags.HasErrors() {
		return diags.Diagnostics()
	}
	logger.Trace("command parsed")
	logger.Tracef("script: %.30s... ", cmd.String())

	if s.Container == nil {
		s.process = cmd
		logger.Tracef("running command: %.30s...", cmd.String())
		if !cfg.Behavior.DryRun {
			err = cmd.Run()

			if err != nil && err.Error() == "signal: terminated" && s.Terminated() {
				logger.Warnf("command terminated with signal: %s", cmd.ProcessState.String())
				err = nil
			}
		} else {
			fmt.Println(cmd.String())
		}
	} else {
		d := s.executeDocker(conductor, evalCtx, cmd, cfg)
		diags.Extend(d)
	}

	if err != nil {
		diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("failed to run command (%s)", s.Identifier()),
			Detail:   err.Error(),
		})
	}

	return diags.Diagnostics()
}

func (s *Stage) executeDocker(conductor *Conductor, evalCtx *hcl.EvalContext, cmd *exec.Cmd, cfg *runnable.Config) hcl.Diagnostics {
	var diags hcl.Diagnostics
	logger := conductor.Logger().WithField("stage", s.Id)

	ctx := conductor.Context()

	image, d := s.hclImage(conductor, evalCtx)
	diags = diags.Extend(d)

	// begin entrypoint evaluation
	entrypoint, d := s.hclEndpoint(conductor, evalCtx)
	diags = diags.Extend(d)

	if diags.HasErrors() {
		return diags
	}

	cli, err := dockerClient.NewClientWithOpts(dockerClient.FromEnv, dockerClient.WithAPIVersionNegotiation())
	if err != nil {
		return diags.Append(&hcl.Diagnostic{
			Severity:    hcl.DiagError,
			Summary:     "could not create docker client",
			Detail:      err.Error(),
			Subject:     s.Container.Image.Range().Ptr(),
			EvalContext: evalCtx,
		})
	}
	defer x.Must(cli.Close())

	// check if image exists
	logger.Debugf("checking if image %s exists", image)
	_, _, err = cli.ImageInspectWithRaw(ctx, image)
	if err != nil {
		logger.Infof("image %s does not exist, pulling...", image)
		reader, err := cli.ImagePull(ctx, image, types.ImagePullOptions{})
		if err != nil {
			return diags.Append(&hcl.Diagnostic{
				Severity:    hcl.DiagError,
				Summary:     "could not pull image",
				Detail:      err.Error(),
				Subject:     s.Container.Image.Range().Ptr(),
				EvalContext: evalCtx,
			})
		}

		pb := ui.NewDockerProgressWriter(reader, logger.Writer(), fmt.Sprintf("pulling image %s", image))
		defer pb.Close()
		defer reader.Close()
		io.Copy(pb, reader)
	}

	logger.Trace("parsing container arguments")
	binds := []string{
		fmt.Sprintf("%s:/workspace", cmd.Dir),
	}

	logger.Trace("parsing container volumes")
	for _, m := range s.Container.Volumes {
		conductor.Eval().Mutex().RLock()
		source, d := m.Source.Value(evalCtx)
		conductor.Eval().Mutex().RUnlock()
		diags = diags.Extend(d)

		conductor.Eval().Mutex().RLock()
		dest, d := m.Destination.Value(evalCtx)
		conductor.Eval().Mutex().RUnlock()
		diags = diags.Extend(d)
		if diags.HasErrors() {
			continue
		}
		binds = append(binds, fmt.Sprintf("%s:%s", source.AsString(), dest.AsString()))
	}
	logger.Tracef("%d diagnostic(s) after parsing container volumes", len(diags.Errs()))
	if diags.HasErrors() {
		return diags
	}

	logger.Trace("dry run check")
	if cfg.Behavior.DryRun {
		fmt.Println(ui.Blue("docker:run.image"), ui.Green(image))
		fmt.Println(ui.Blue("docker:run.workdir"), ui.Green("/workspace"))
		fmt.Println(ui.Blue("docker:run.volume"), ui.Green(cmd.Dir+":/workspace"))
		fmt.Println(ui.Blue("docker:run.stdin"), ui.Green(s.Container.Stdin))
		fmt.Println(ui.Blue("docker:run.args"), ui.Green(cmd.String()))
		return diags
	}

	logger.Trace("parsing container ports")
	exposedPorts, bindings, d := s.Container.Ports.Nat(conductor, evalCtx)
	diags = diags.Extend(d)
	if diags.HasErrors() {
		return diags
	}

	logger.Trace("creating container")
	resp, err := cli.ContainerCreate(conductor.Context(), &dockerContainer.Config{
		Image:      image,
		Cmd:        cmd.Args,
		WorkingDir: "/workspace",
		Volumes: map[string]struct{}{
			"/workspace": {},
		},
		Tty:          true,
		AttachStdout: true,
		AttachStderr: true,
		AttachStdin:  s.Container.Stdin,
		OpenStdin:    s.Container.Stdin,
		StdinOnce:    s.Container.Stdin,
		Entrypoint:   entrypoint,
		Env:          cmd.Env,
		ExposedPorts: exposedPorts,
		// User: s.Container.User,
	}, &dockerContainer.HostConfig{
		Binds:        binds,
		PortBindings: bindings,
	}, nil, nil, "")
	if err != nil {
		return diags.Append(&hcl.Diagnostic{
			Severity:    hcl.DiagError,
			Summary:     "could not create container",
			Detail:      err.Error(),
			Subject:     s.Container.Image.Range().Ptr(),
			EvalContext: evalCtx,
		})
	}

	logger.Trace("starting container")
	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "could not start container",
			Detail:   err.Error(),
			Subject:  s.Container.Image.Range().Ptr(),
		})
	}
	s.ContainerId = resp.ID

	logger.Trace("getting container metadata for log retrieval")
	container, err := cli.ContainerInspect(ctx, resp.ID)
	if err != nil {
		panic(err)
	}

	logger.Trace("getting container logs")
	responseBody, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{
		ShowStdout: true, ShowStderr: true,
		Follow: true,
	})
	if err != nil {
		return diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "could not get container logs",
			Detail:   err.Error(),
			Subject:  s.Container.Image.Range().Ptr(),
		})
	}
	defer responseBody.Close()

	logger.Tracef("copying container logs on container: %s", resp.ID)
	if container.Config.Tty {
		_, err = io.Copy(logger.Writer(), responseBody)
	} else {
		_, err = stdcopy.StdCopy(logger.Writer(), logger.WriterLevel(logrus.WarnLevel), responseBody)
	}

	logger.Trace("waiting for container to finish")
	if err != nil && err != io.EOF {
		if errors.Is(err, context.Canceled) {
			return diags
		}
		diags = diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "failed to copy container logs",
			Detail:   err.Error(),
			Subject:  s.Container.Image.Range().Ptr(),
		})
	}

	logger.Tracef("removing container with id: %s", resp.ID)
	err = cli.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{
		RemoveVolumes: true,
	})
	if err != nil {
		diags = diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "failed to remove container",
			Detail:   err.Error(),
			Subject:  s.Container.Image.Range().Ptr(),
		})
		return diags
	}

	logger.Tracef("%d diagnostic(s) after removing container", len(diags.Errs()))
	return diags
}

func (s *Stage) parseEnvironmentVariables(conductor *Conductor, evalCtx *hcl.EvalContext) (map[string]cty.Value, hcl.Diagnostics) {
	logger := conductor.Logger().WithField("stage", s.Id)
	var diags hcl.Diagnostics
	logger.Debug("evaluating environment variables")
	var environment map[string]cty.Value
	environment = make(map[string]cty.Value)
	for _, env := range s.Environment {
		conductor.Eval().Mutex().RLock()
		v, d := env.Value.Value(evalCtx)
		conductor.Eval().Mutex().RUnlock()

		diags = diags.Extend(d)
		if v.IsNull() {
			diags = diags.Append(&hcl.Diagnostic{
				Severity:    hcl.DiagError,
				Summary:     "invalid environment variable",
				Detail:      fmt.Sprintf("environment variable %s is null", env.Name),
				EvalContext: evalCtx,
				Subject:     env.Value.Range().Ptr(),
			})
		} else if v.Type() != cty.String {
			diags = diags.Append(&hcl.Diagnostic{
				Severity:    hcl.DiagError,
				Summary:     "invalid environment variable",
				Detail:      fmt.Sprintf("environment variable %s is not a string", env.Name),
				EvalContext: evalCtx,
				Subject:     env.Value.Range().Ptr(),
			})
		} else {
			environment[env.Name] = v
		}
	}
	return environment, diags
}

func (s *Stage) parseExecCommand(conductor *Conductor, evalCtx *hcl.EvalContext, cfg *runnable.Config, envStrings []string, outputBuffer io.Writer) (*exec.Cmd, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	logger := conductor.Logger().WithField("stage", s.Id)

	logger.Trace("evaluating script value")
	conductor.Eval().Mutex().RLock()
	script, d := s.Script.Value(evalCtx)
	conductor.Eval().Mutex().RUnlock()

	if d.HasErrors() && cfg.Behavior.DryRun {
		script = cty.StringVal(ui.Italic(ui.Yellow("(will be evaluated later)")))
	} else {
		diags = diags.Extend(d)
	}

	logger.Trace("evaluating shell value")
	conductor.Eval().Mutex().RLock()
	shellRaw, d := s.Shell.Value(evalCtx)
	conductor.Eval().Mutex().RUnlock()

	shell := ""
	if d.HasErrors() {
		diags = diags.Extend(d)
	} else {
		if shellRaw.IsNull() {
			shell = "bash"
		} else {
			shell = shellRaw.AsString()
		}
	}

	logger.Trace("evaluating args value")
	conductor.Eval().Mutex().RLock()
	args, d := s.Args.Value(evalCtx)
	conductor.Eval().Mutex().RUnlock()
	diags = diags.Extend(d)

	cmdHcl, d := s.parseCommand(evalCtx, shell, script, args)
	diags = diags.Extend(d)
	if diags.HasErrors() {
		return nil, diags
	}

	dir := cfg.Paths.Cwd

	conductor.Eval().Mutex().RLock()
	dirParsed, d := s.Dir.Value(evalCtx)
	conductor.Eval().Mutex().RUnlock()

	if d.HasErrors() {
		diags = diags.Extend(d)
	} else {
		if !dirParsed.IsNull() && dirParsed.AsString() != "" {
			dir = dirParsed.AsString()
		}
		if !filepath.IsAbs(dir) {
			dir = filepath.Join(cfg.Paths.Cwd, dir)
		}
		if cfg.Behavior.DryRun {
			fmt.Println(ui.Blue("cd"), dir)
		}
	}

	cmd := exec.CommandContext(conductor.Context(), cmdHcl.command, cmdHcl.args...)
	cmd.Stdout = io.MultiWriter(logger.Writer(), outputBuffer)
	cmd.Stderr = io.MultiWriter(logger.Writer(), outputBuffer)
	cmd.Env = append(os.Environ(), envStrings...)
	cmd.Dir = dir
	return cmd, diags
}

type command struct {
	args    []string
	command string

	isEmpty bool
}

func (s *Stage) parseCommand(evalCtx *hcl.EvalContext, shell string, script cty.Value, args cty.Value) (command, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	runArgs := make([]string, 0)
	runCommand := shell

	// emptyCommands - specifies if both args and scripts were unset
	emptyCommands := false
	if script.Type() == cty.String {
		if !script.IsKnown() {
			diags = diags.Append(&hcl.Diagnostic{
				Severity:    hcl.DiagError,
				Summary:     "invalid script",
				Detail:      fmt.Sprintf("script is not a valid string"),
				Subject:     s.Script.Range().Ptr(),
				EvalContext: evalCtx,
			})
			return command{}, diags
		}
		if shell == "bash" {
			runArgs = append(runArgs, "-e", "-u", "-c", script.AsString())
		} else if shell == "sh" {
			runArgs = append(runArgs, "-e", "-c", script.AsString())
		} else {
			runArgs = append(runArgs, script.AsString())
		}
	} else if !args.IsNull() && len(args.AsValueSlice()) != 0 {
		runCommand = args.AsValueSlice()[0].AsString()
		for i, a := range args.AsValueSlice() {
			if i == 0 {
				continue
			}
			runArgs = append(runArgs, a.AsString())
		}
	} else if s.Container == nil {
		// if the container is not null, we may rely on internal args or entrypoint scripts
		diags = diags.Append(&hcl.Diagnostic{
			Severity:    hcl.DiagError,
			Summary:     "No commands specified",
			Detail:      "Either script or args must be specified",
			Subject:     s.Script.Range().Ptr(),
			EvalContext: evalCtx,
		})
	} else {
		emptyCommands = true
	}
	return command{
		args:    runArgs,
		command: runCommand,
		isEmpty: emptyCommands,
	}, diags
}

func (s *Stage) processEnvironmentVariables(conductor *Conductor, environment map[string]cty.Value, cfg *runnable.Config, tmpDir string, paramsGo map[string]cty.Value) []string {
	logger := conductor.Logger().WithField("stage", s.Id)
	envStrings := make([]string, len(environment))
	envCounter := 0
	for k, v := range environment {
		envParsed := fmt.Sprintf("%s=%s", k, v.AsString())
		if cfg.Behavior.DryRun {
			fmt.Println(ui.Blue("export"), envParsed)
		}

		envStrings[envCounter] = envParsed
		envCounter = envCounter + 1
	}
	togomakEnvExport := fmt.Sprintf("%s=%s", meta.OutputEnvVar, filepath.Join(tmpDir, meta.OutputEnvFile))
	logger.Tracef("exporting %s", togomakEnvExport)
	envStrings = append(envStrings, togomakEnvExport)

	if s.Use != nil && s.Use.Parameters != nil {
		for k, v := range paramsGo {
			envParsed := fmt.Sprintf("%s%s=%s", TogomakParamEnvVarPrefix, k, v.AsString())
			if cfg.Behavior.DryRun {
				fmt.Println(ui.Blue("export"), envParsed)
			}

			envStrings = append(envStrings, envParsed)
		}
	}
	return envStrings
}

func (s *Stage) executePreHooks(conductor *Conductor, status runnable.StatusType, options ...runnable.Option) hcl.Diagnostics {
	logger := conductor.Logger().WithField("stage", s.Id)
	var diags hcl.Diagnostics
	logger.Debugf("running pre hooks")
	hookOpts := []runnable.Option{
		runnable.WithStatus(status),
		runnable.WithHook(),
		runnable.WithParent(runnable.ParentConfig{Name: s.Name, Id: s.Id}),
	}
	hookOpts = append(hookOpts, options...)
	diags = diags.Extend(s.BeforeRun(conductor, hookOpts...))
	logger.Debugf("finished running pre hooks")
	return diags
}

func (s *Stage) hclImage(conductor *Conductor, evalCtx *hcl.EvalContext) (image string, diags hcl.Diagnostics) {
	conductor.Eval().Mutex().RLock()
	imageRaw, d := s.Container.Image.Value(evalCtx)
	conductor.Eval().Mutex().RUnlock()

	if d.HasErrors() {
		diags = diags.Extend(d)
	} else if imageRaw.Type() != cty.String {
		diags = diags.Append(&hcl.Diagnostic{
			Severity:    hcl.DiagError,
			Summary:     "image must be a string",
			Detail:      fmt.Sprintf("the provided image, was not recognized as a valid string. received image='''%s'''", imageRaw),
			Subject:     s.Container.Image.Range().Ptr(),
			EvalContext: evalCtx,
		})
	} else {
		image = imageRaw.AsString()
	}
	return image, diags
}

func (s *Stage) hclEndpoint(conductor *Conductor, evalCtx *hcl.EvalContext) ([]string, hcl.Diagnostics) {
	var diags hcl.Diagnostics

	conductor.Eval().Mutex().RLock()
	entrypointRaw, d := s.Container.Entrypoint.Value(evalCtx)
	conductor.Eval().Mutex().RUnlock()

	var entrypoint []string
	if d.HasErrors() {
		diags = diags.Extend(d)
	} else if entrypointRaw.IsNull() {
		entrypoint = nil
	} else if !entrypointRaw.CanIterateElements() {
		diags = diags.Append(&hcl.Diagnostic{
			Severity:    hcl.DiagError,
			Summary:     "entrypoint must be a list of strings",
			Detail:      fmt.Sprintf("the provided entrypoint, was not recognized as a valid string. received entrypoint='''%s'''", entrypointRaw),
			Subject:     s.Container.Entrypoint.Range().Ptr(),
			EvalContext: evalCtx,
		})
	} else {
		v := entrypointRaw.AsValueSlice()
		for _, e := range v {
			entrypoint = append(entrypoint, e.AsString())
		}
	}
	return entrypoint, diags
}

func (s *Stage) CanRun(conductor *Conductor, options ...runnable.Option) (ok bool, diags hcl.Diagnostics) {
	logger := conductor.Logger().WithField("stage", s.Id)
	logger.Debugf("checking if stage.%s can run", s.Id)
	evalCtx := conductor.Eval().Context()

	cfg := runnable.NewConfig(options...)

	paramsGo := map[string]cty.Value{}

	evalCtx = evalCtx.NewChild()
	name := s.Name
	id := s.Id
	if cfg.Parent != nil {
		name = cfg.Parent.Name
		id = cfg.Parent.Id
	}

	evalCtx.Variables = map[string]cty.Value{
		"this": cty.ObjectVal(map[string]cty.Value{
			"name":   cty.StringVal(name),
			"id":     cty.StringVal(id),
			"hook":   cty.BoolVal(cfg.Hook),
			"status": cty.StringVal(string(cfg.Status.Status)),
			"output": cty.StringVal(cfg.Status.Output),
		}),
		"param": cty.ObjectVal(paramsGo),
	}

	if s.Use != nil && s.Use.Parameters != nil {
		conductor.Eval().Mutex().RLock()
		parameters, d := s.Use.Parameters.Value(evalCtx)
		conductor.Eval().Mutex().RUnlock()

		diags = diags.Extend(d)
		if !parameters.IsNull() {
			for k, v := range parameters.AsValueMap() {
				paramsGo[k] = v
			}
		}
	}

	conductor.Eval().Mutex().RLock()
	v, d := s.Condition.Value(evalCtx)
	conductor.Eval().Mutex().RUnlock()
	if d.HasErrors() {
		return false, diags.Extend(d)
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
