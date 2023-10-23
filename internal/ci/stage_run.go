package ci

import (
	"fmt"
	"github.com/alessio/shellescape"
	"github.com/srevinsaju/togomak/v1/internal/blocks"
	"github.com/srevinsaju/togomak/v1/internal/dg"
	"github.com/srevinsaju/togomak/v1/internal/x"
	"sync"

	"github.com/google/uuid"
	"github.com/hashicorp/hcl/v2"
	"github.com/imdario/mergo"
	"github.com/srevinsaju/togomak/v1/internal/c"
	"github.com/srevinsaju/togomak/v1/internal/global"
	"github.com/srevinsaju/togomak/v1/internal/meta"
	"github.com/srevinsaju/togomak/v1/internal/runnable"
	"github.com/srevinsaju/togomak/v1/internal/ui"
	"github.com/zclconf/go-cty/cty"
	"os"
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

	tmpDir := global.TempDir()

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
		if !strings.HasSuffix(sourceEvaluated, ".hcl") {
			sourceEvaluated = filepath.Join(sourceEvaluated, meta.ConfigFileName)
		}
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
		return false
	})
	wg.Wait()
	return safeDg.Diagnostics()

}

func (s *Stage) run(conductor *Conductor, evalCtx *hcl.EvalContext, options ...runnable.Option) (diags hcl.Diagnostics) {
	cfg := runnable.NewConfig(options...)
	c := s.newStageConductor(conductor, cfg)

	logger := conductor.Logger().WithField("stage", s.Id)
	status := runnable.StatusRunning

	defer func() {
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
			runnable.WithParent(runnable.ParentConfig{Name: s.Name, Id: s.Id}),
		}
		hookOpts = append(hookOpts, options...)
		diags = diags.Extend(s.AfterRun(conductor, hookOpts...))
		logger.Debug("finished running post hooks")
	}()

	d := s.executePreHooks(conductor, status, options...)
	diags = diags.Extend(d)

	c = c.
		expandParams().
		expandSelf().
		expandEach().
		parseEnvironment().
		processEnvironment().
		parseExecCommand()

	if c.diags.HasErrors() {
		return diags.Extend(c.diags)
	}

	if s.Container == nil {
		c = c.runUsingShell()
	} else {
		c = c.runUsingDocker()
	}
	if c.diags.HasErrors() {
		return diags.Extend(c.diags)
	}

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
