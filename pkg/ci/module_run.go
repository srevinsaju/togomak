package ci

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-getter"
	"github.com/hashicorp/hcl/v2"
	"github.com/srevinsaju/togomak/v1/pkg/global"
	"github.com/srevinsaju/togomak/v1/pkg/runnable"
	"github.com/srevinsaju/togomak/v1/pkg/ui"
	"github.com/srevinsaju/togomak/v1/pkg/x"
	"github.com/zclconf/go-cty/cty"
)

func (m *Module) Prepare(ctx context.Context, skip bool, overridden bool) hcl.Diagnostics {
	logger := m.Logger()
	// show some user-friendly output on the details of the stage about to be run

	var id string
	identifier := fmt.Sprintf("%s.%s", ModuleBlock, m.Id)
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

func (m *Module) Run(ctx context.Context, options ...runnable.Option) (diags hcl.Diagnostics) {
	//TODO implement me
	logger := m.Logger()
	logger.Debugf("running %s", x.RenderBlock(ModuleBlock, m.Id))
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

	src := v.AsString()
	get := &getter.Client{
		Ctx: ctx,
		Src: src,
		Dst: "",
		Pwd: "",
		Dir: true,
	}
	get.Get()
	return diags
}

func (m *Module) CanRun(ctx context.Context, options ...runnable.Option) (ok bool, diags hcl.Diagnostics) {
	logger := m.Logger()
	logger.Debugf("checking if %s can run", x.RenderBlock(ModuleBlock, m.Id))
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
