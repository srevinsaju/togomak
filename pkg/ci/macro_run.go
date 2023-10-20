package ci

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/srevinsaju/togomak/v1/pkg/blocks"
	"github.com/srevinsaju/togomak/v1/pkg/global"
	"github.com/srevinsaju/togomak/v1/pkg/runnable"
	"github.com/zclconf/go-cty/cty"
)

const (
	SourceTypeGit = "git"
)

func (m *Macro) Prepare(conductor *Conductor, skip bool, overridden bool) hcl.Diagnostics {
	return nil // no-op
}

func (m *Macro) Run(conductor *Conductor, options ...runnable.Option) (diags hcl.Diagnostics) {
	// _ := ctx.Value(TogomakContextHclDiagWriter).(hcl.DiagnosticWriter)
	logger := m.Logger()
	logger.Tracef("running %s.%s", blocks.MacroBlock, m.Id)
	hclContext := global.HclEvalContext()

	// region: mutating the data map
	// TODO: move it to a dedicated helper function

	global.MacroBlockEvalContextMutex.Lock()

	global.EvalContextMutex.RLock()
	macro := hclContext.Variables[blocks.MacroBlock]

	var macroMutated map[string]cty.Value
	if macro.IsNull() {
		macroMutated = make(map[string]cty.Value)
	} else {
		macroMutated = macro.AsValueMap()
	}
	// -> update r.Value accordingly
	f, d := m.Files.Value(hclContext)
	global.EvalContextMutex.RUnlock()

	if d != nil {
		diags = diags.Extend(d)
	}
	macroMutated[m.Id] = cty.ObjectVal(map[string]cty.Value{
		"files": f,
	})

	global.EvalContextMutex.Lock()
	hclContext.Variables[blocks.MacroBlock] = cty.ObjectVal(macroMutated)
	global.EvalContextMutex.Unlock()

	global.MacroBlockEvalContextMutex.Unlock()
	// endregion

	return diags
}

func (m *Macro) CanRun(conductor *Conductor, options ...runnable.Option) (ok bool, diags hcl.Diagnostics) {
	return false, diags
}

func (m *Macro) Terminated() bool {
	return true
}
