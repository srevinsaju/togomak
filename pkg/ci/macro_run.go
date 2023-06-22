package ci

import (
	"context"
	"github.com/hashicorp/hcl/v2"
	"github.com/sirupsen/logrus"
	"github.com/srevinsaju/togomak/v1/pkg/c"
	"github.com/zclconf/go-cty/cty"
)

const (
	SourceTypeGit = "git"
)

func (m *Macro) Prepare(ctx context.Context, skip bool, overridden bool) hcl.Diagnostics {
	return nil // no-op
}

func (m *Macro) Run(ctx context.Context) hcl.Diagnostics {
	// _ := ctx.Value(TogomakContextHclDiagWriter).(hcl.DiagnosticWriter)
	logger := ctx.Value(c.TogomakContextLogger).(*logrus.Logger).WithField(DataBlock, m.Id)
	logger.Tracef("running %s.%s", MacroBlock, m.Id)
	hclContext := ctx.Value(c.TogomakContextHclEval).(*hcl.EvalContext)
	var diags hcl.Diagnostics

	// region: mutating the data map
	// TODO: move it to a dedicated helper function
	macro := hclContext.Variables[MacroBlock]
	var macroMutated map[string]cty.Value
	if macro.IsNull() {
		macroMutated = make(map[string]cty.Value)
	} else {
		macroMutated = macro.AsValueMap()
	}

	// -> update r.Value accordingly
	f, d := m.Files.Value(hclContext)
	if d != nil {
		diags = diags.Extend(d)
	}
	macroMutated[m.Id] = cty.ObjectVal(map[string]cty.Value{
		"files": f,
	})
	hclContext.Variables[MacroBlock] = cty.ObjectVal(macroMutated)
	// endregion

	return diags
}

func (m *Macro) CanRun(ctx context.Context) (bool, hcl.Diagnostics) {
	return true, nil
}

func (m *Macro) Terminated() bool {
	return true
}
