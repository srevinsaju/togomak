package ci

import (
	"fmt"
	"github.com/hashicorp/go-getter"
	"github.com/hashicorp/hcl/v2"
	"github.com/srevinsaju/togomak/v1/internal/blocks"
	"github.com/srevinsaju/togomak/v1/internal/global"
	"github.com/srevinsaju/togomak/v1/internal/runnable"
	"github.com/srevinsaju/togomak/v1/internal/ui"
	"github.com/srevinsaju/togomak/v1/internal/x"
	"github.com/zclconf/go-cty/cty"
	"path/filepath"
)

func (m *Module) Prepare(conductor *Conductor, skip bool, overridden bool) hcl.Diagnostics {
	logger := m.Logger()
	// show some user-friendly output on the details of the stage about to be run

	var id string
	identifier := fmt.Sprintf("%s.%s", blocks.ModuleBlock, m.Id)
	if !skip {
		id = ui.Blue(identifier)
	} else {
		id = fmt.Sprintf("%s %s", ui.Yellow(identifier), ui.Grey("(skipped)"))
	}
	if overridden {
		id = fmt.Sprintf("%s %s", id, ui.Bold("(overriden)"))
	}
	logger.Infof("[%s] %s", ui.Plus, id)
	return nil
}

func (m *Module) Run(conductor *Conductor, options ...runnable.Option) (diags hcl.Diagnostics) {
	//TODO implement me
	logger := m.Logger()
	logger.Debugf("running %s", x.RenderBlock(blocks.ModuleBlock, m.Id))
	evalCtx := global.HclEvalContext()
	evalCtx = evalCtx.NewChild()

	cfg := runnable.NewConfig(options...)

	evalCtx.Variables = map[string]cty.Value{
		"this": cty.ObjectVal(map[string]cty.Value{
			"id":     cty.StringVal(m.Id),
			"status": cty.StringVal(string(cfg.Status.Status)),
		}),
	}

	global.EvalContextMutex.RLock()
	v, d := m.Source.Value(evalCtx)
	global.EvalContextMutex.RUnlock()
	if d.HasErrors() {
		return diags.Extend(d)
	}
	if v.Type() != cty.String {
		return diags.Append(&hcl.Diagnostic{
			Severity:    hcl.DiagError,
			Summary:     "source must be a string",
			Detail:      fmt.Sprintf("source must be a string, got %s", v.Type().FriendlyName()),
			Subject:     m.Source.Range().Ptr(),
			EvalContext: evalCtx,
		})
	}

	paths := cfg.Paths
	src := v.AsString()
	get := &getter.Client{
		Ctx: conductor.Context(),
		Src: src,
		Dst: filepath.Join(global.TempDir(), "modules", m.Id),
		Pwd: paths.Cwd,
		Dir: true,
	}
	err := get.Get()
	if err != nil {
		return diags.Append(&hcl.Diagnostic{
			Severity:    hcl.DiagError,
			Summary:     "failed to download source",
			Detail:      err.Error(),
			Subject:     m.Source.Range().Ptr(),
			EvalContext: evalCtx,
		})
	}

	childConductor := conductor.Child()

	// parse the config file
	pipe, hclDiags := Read(conductor.Config.Paths, conductor.Parser)
	if hclDiags.HasErrors() {
		logger.Fatal(conductor.DiagWriter.WriteDiagnostics(hclDiags))
	}

	return pipe.Run(conductor)

	return diags
}

func (m *Module) CanRun(conductor *Conductor, options ...runnable.Option) (ok bool, diags hcl.Diagnostics) {
	logger := m.Logger()
	logger.Debugf("checking if %s can run", x.RenderBlock(blocks.ModuleBlock, m.Id))
	evalCtx := global.HclEvalContext()
	evalCtx = evalCtx.NewChild()

	cfg := runnable.NewConfig(options...)

	evalCtx.Variables = map[string]cty.Value{
		"this": cty.ObjectVal(map[string]cty.Value{
			"id":     cty.StringVal(m.Id),
			"status": cty.StringVal(string(cfg.Status.Status)),
		}),
	}

	global.EvalContextMutex.RLock()
	v, d := m.Condition.Value(evalCtx)
	global.EvalContextMutex.RUnlock()
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

func (m *Module) Terminated() bool {
	//TODO implement me
	panic("implement me")
}
