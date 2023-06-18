package ci

import (
	"context"
	"github.com/hashicorp/hcl/v2"
	"github.com/sirupsen/logrus"
	"github.com/srevinsaju/togomak/v1/pkg/c"
	"github.com/srevinsaju/togomak/v1/pkg/diag"
	"github.com/zclconf/go-cty/cty"
)

const (
	SourceTypeGit = "git"
)

func (m *Macro) Prepare(ctx context.Context, skip bool, overridden bool) diag.Diagnostics {
	return nil // no-op
}

func (m *Macro) Run(ctx context.Context) diag.Diagnostics {
	// _ := ctx.Value(TogomakContextHclDiagWriter).(hcl.DiagnosticWriter)
	logger := ctx.Value(c.TogomakContextLogger).(*logrus.Logger).WithField(DataBlock, m.Id)
	logger.Tracef("running %s.%s", MacroBlock, m.Id)
	hclContext := ctx.Value(c.TogomakContextHclEval).(*hcl.EvalContext)
	hcDiagWriter := ctx.Value(c.TogomakContextHclDiagWriter).(hcl.DiagnosticWriter)
	var hcDiags hcl.Diagnostics
	var diags diag.Diagnostics

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
		hcDiags = hcDiags.Extend(d)
	}
	macroMutated[m.Id] = cty.ObjectVal(map[string]cty.Value{
		"files": f,
	})
	hclContext.Variables[MacroBlock] = cty.ObjectVal(macroMutated)
	// endregion

	if hcDiags.HasErrors() {
		diags = diags.ExtendHCLDiagnostics(hcDiags, hcDiagWriter, m.Identifier())
	}

	if diags.HasErrors() {
		return diags
	}
	return nil
}

func (m *Macro) CanRun(ctx context.Context) (bool, diag.Diagnostics) {
	return true, nil
}
