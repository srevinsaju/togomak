package ci

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/srevinsaju/togomak/v1/internal/blocks"
	"github.com/srevinsaju/togomak/v1/internal/global"
	"github.com/srevinsaju/togomak/v1/internal/runnable"
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
	logger := conductor.Logger().WithField("macro", m.Id)
	logger.Tracef("running %s.%s", blocks.MacroBlock, m.Id)
	evalContext := conductor.Eval().Context()

	// region: mutating the data map
	// TODO: move it to a dedicated helper function

	global.MacroBlockEvalContextMutex.Lock()

	conductor.Eval().Mutex().RLock()
	macro := evalContext.Variables[blocks.MacroBlock]

	var macroMutated map[string]cty.Value
	if macro.IsNull() {
		macroMutated = make(map[string]cty.Value)
	} else {
		macroMutated = macro.AsValueMap()
	}
	// -> update r.Value accordingly
	f, d := m.Files.Value(evalContext)
	conductor.Eval().Mutex().RUnlock()

	if d != nil {
		diags = diags.Extend(d)
	}
	macroMutated[m.Id] = cty.ObjectVal(map[string]cty.Value{
		"files": f,
	})

	conductor.Eval().Mutex().Lock()
	evalContext.Variables[blocks.MacroBlock] = cty.ObjectVal(macroMutated)
	conductor.Eval().Mutex().Unlock()

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
