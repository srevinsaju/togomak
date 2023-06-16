package ci

import (
	"context"
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/sirupsen/logrus"
	dataBlock "github.com/srevinsaju/togomak/v1/pkg/blocks/data"
	"github.com/srevinsaju/togomak/v1/pkg/c"
	"github.com/srevinsaju/togomak/v1/pkg/diag"
	"github.com/zclconf/go-cty/cty"
	"sync"
)

func (d Data) Prepare(ctx context.Context, skip bool, overridden bool) diag.Diagnostics {
	return nil // no-op
}

func (d Data) Run(ctx context.Context) diag.Diagnostics {
	// _ := ctx.Value(TogomakContextHclDiagWriter).(hcl.DiagnosticWriter)
	logger := ctx.Value(c.TogomakContextLogger).(*logrus.Logger).WithField(DataBlock, d.Id)
	logger.Debugf("running %s.%s.%s", DataBlock, d.Provider, d.Id)
	hclContext := ctx.Value(c.TogomakContextHclEval).(*hcl.EvalContext)
	hcDiagWriter := ctx.Value(c.TogomakContextHclDiagWriter).(hcl.DiagnosticWriter)
	muData := ctx.Value(c.TogomakContextMutexData).(*sync.Mutex)
	var hcDiags hcl.Diagnostics
	var diags diag.Diagnostics

	// region: mutating the data map
	// TODO: move it to a dedicated helper function

	// -> update r.Value accordingly
	var validProvider bool
	var value string
	var attr map[string]cty.Value
	for _, pr := range dataBlock.DefaultProviders {
		if pr.Name() == d.Provider {
			validProvider = true
			provide := pr.New()
			provide.SetContext(ctx)
			diags = diags.Extend(provide.DecodeBody(d.Body))
			value = provide.Value(ctx)
			attr = provide.Attributes(ctx)
			break
		}
	}
	if !validProvider || diags.HasErrors() {
		diags = diags.Append(diag.Diagnostic{
			Severity: diag.SeverityInvalid,
			Summary:  fmt.Sprintf("invalid provider %s", d.Provider),
			Detail:   fmt.Sprintf("built-in providers are %v", dataBlock.DefaultProviders),
		})
		return diags
	}
	m := make(map[string]cty.Value)
	m["value"] = cty.StringVal(value)
	for k, v := range attr {
		m[k] = v
	}

	muData.Lock()
	data := hclContext.Variables[DataBlock]
	var dataMutated map[string]cty.Value
	if data.IsNull() {
		dataMutated = make(map[string]cty.Value)
	} else {
		dataMutated = data.AsValueMap()
	}
	provider := dataMutated[d.Provider]
	var providerMutated map[string]cty.Value
	if provider.IsNull() {
		providerMutated = make(map[string]cty.Value)
	} else {
		providerMutated = provider.AsValueMap()
	}
	providerMutated[d.Id] = cty.ObjectVal(m)
	dataMutated[d.Provider] = cty.ObjectVal(providerMutated)
	hclContext.Variables[DataBlock] = cty.ObjectVal(dataMutated)
	muData.Unlock()
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

	v := d.Variables()
	logger.Debug(fmt.Sprintf("data.%s variables: %v", d.Id, v))
	return nil
}

func (d Data) CanRun(ctx context.Context) (bool, diag.Diagnostics) {
	return true, nil
}
