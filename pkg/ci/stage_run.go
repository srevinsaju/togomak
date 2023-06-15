package ci

import (
	"context"
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/sirupsen/logrus"
	"github.com/srevinsaju/togomak/v1/pkg/c"
	"github.com/srevinsaju/togomak/v1/pkg/diag"
	"github.com/srevinsaju/togomak/v1/pkg/ui"
	"github.com/zclconf/go-cty/cty"
	"os"
	"os/exec"
)

func (s *Stage) Prepare(ctx context.Context, skip bool, overridden bool) {
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

	paramsGo := map[string]cty.Value{}
	if s.Use != nil && s.Use.Parameters != nil {
		parameters, d := s.Use.Parameters.Value(evalCtx)
		hclDiags = hclDiags.Extend(d)

		for k, v := range parameters.AsValueMap() {
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

	runArgs := make([]string, 0)
	runCommand := "sh"
	if script.Type() == cty.String {
		runArgs = append(runArgs, "-c", script.AsString())
	} else if args.Type() == cty.List(cty.String) {
		for _, a := range args.AsValueSlice() {
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

		for k, v := range parameters.AsValueMap() {
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
