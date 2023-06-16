package ci

import (
	"context"
	"github.com/hashicorp/hcl/v2"
	"github.com/sirupsen/logrus"
	"github.com/srevinsaju/togomak/v1/pkg/c"
	"github.com/srevinsaju/togomak/v1/pkg/diag"
	"github.com/zclconf/go-cty/cty"
	"sync"
)

func (l *Local) Run(ctx context.Context) diag.Diagnostics {

	logger := ctx.Value(c.TogomakContextLogger).(*logrus.Logger).WithField(LocalBlock, l.Key)
	logger.Debugf("running %s.%s", LocalBlock, l.Key)
	hclContext := ctx.Value(c.TogomakContextHclEval).(*hcl.EvalContext)
	hcDiagWriter := ctx.Value(c.TogomakContextHclDiagWriter).(hcl.DiagnosticWriter)
	muLocals := ctx.Value(c.TogomakContextMutexLocals).(*sync.Mutex)
	var hcDiags hcl.Diagnostics
	var diags diag.Diagnostics

	// region: mutating the data map
	// TODO: move it to a dedicated helper function
	muLocals.Lock()
	locals := hclContext.Variables[LocalBlock]
	var localMutated map[string]cty.Value
	if locals.IsNull() {
		localMutated = make(map[string]cty.Value)
	} else {
		localMutated = locals.AsValueMap()
	}
	v, d := l.Value.Value(hclContext)
	hcDiags = hcDiags.Extend(d)
	localMutated[l.Key] = v
	hclContext.Variables[LocalBlock] = cty.ObjectVal(localMutated)
	muLocals.Unlock()
	// endregion

	if hcDiags.HasErrors() {
		err := hcDiagWriter.WriteDiagnostics(hcDiags)
		if err != nil {
			diags = diags.Append(diag.Diagnostic{
				Severity: diag.SeverityError,
				Summary:  "failed to write HCL diagnostics",
				Detail:   err.Error(),
			})
		}
		diags = diags.Append(diag.Diagnostic{
			Severity: diag.SeverityError,
			Summary:  "failed to evaluate data block",
			Detail:   hcDiags.Error(),
		})
	}

	if diags.HasErrors() {
		return diags
	}
	return nil
}

func (l *Local) CanRun(ctx context.Context) (bool, diag.Diagnostics) {
	return true, nil
}

func (l *Local) Prepare(ctx context.Context, skip bool, overridden bool) diag.Diagnostics {
	return nil
}
