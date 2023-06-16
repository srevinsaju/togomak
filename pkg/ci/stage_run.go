package ci

import (
	"context"
	"fmt"
	"github.com/alessio/shellescape"
	"github.com/hashicorp/hcl/v2"
	"github.com/imdario/mergo"
	"github.com/sirupsen/logrus"
	"github.com/srevinsaju/togomak/v1/pkg/c"
	"github.com/srevinsaju/togomak/v1/pkg/diag"
	"github.com/srevinsaju/togomak/v1/pkg/ui"
	"github.com/zclconf/go-cty/cty"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
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
	logger.Debugf("running %s.%s.%s", StageBlock, s.Id, MacroBlock)

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
				if filepath.Base(fName) == "togomak.hcl" {
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
	container := s.Container

	if hclDiags.HasErrors() {
		err := hclDgWriter.WriteDiagnostics(hclDiags)
		if err != nil {
			diags = diags.Append(diag.Diagnostic{
				Severity: diag.SeverityError,
				Summary:  "failed to write HCL diagnostics",
				Detail:   err.Error(),
			})
		}
		diags = diags.Append(diag.Diagnostic{
			Severity: diag.SeverityError,
			Summary:  "failed to evaluate HCL",
			Detail:   hclDiags.Error(),
			Source:   fmt.Sprintf("stage.%s:run", s.Id),
		})
	}

	if hclDiags.HasErrors() || diags.HasErrors() {
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
	if script.Type() == cty.String {
		runArgs = append(runArgs, "-c", script.AsString())
	} else if args.Type() == cty.List(cty.String) {
		runCommand = args.AsValueSlice()[0].AsString()
		for i, a := range args.AsValueSlice() {
			if i == 0 {
				continue
			}
			runArgs = append(runArgs, a.AsString())
		}
	} else {
		diags = diags.Append(diag.Diagnostic{
			Severity: diag.SeverityError,
			Summary:  "script or args must be a string",
			Detail:   fmt.Sprintf("the provided script or args, was not recognized as a valid string. received script='''%s''', args='''%s'''", script, args),
		})
		return diags
	}
	logger.Tracef("container=%s", container)
	cmd := exec.CommandContext(ctx, runCommand, runArgs...)
	cmd.Stdout = logger.Writer()
	cmd.Stderr = logger.WriterLevel(logrus.WarnLevel)
	cmd.Env = append(envStrings, os.Environ()...)
	s.process = cmd

	logger.Trace("running command:", cmd.String())

	if !isDryRun {
		err = cmd.Run()
	} else {
		fmt.Println(cmd.String())
	}

	if err != nil {
		diags = diags.Append(diag.Diagnostic{
			Severity: diag.SeverityError,
			Summary:  "failed to run command",
			Detail:   err.Error(),
			Source:   fmt.Sprintf("stage.%s", s.Id),
		})
		return diags
	}

	return nil
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
