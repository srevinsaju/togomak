package ci

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	dataBlock "github.com/srevinsaju/togomak/v1/internal/blocks/data"
	"github.com/srevinsaju/togomak/v1/internal/global"
	"github.com/srevinsaju/togomak/v1/internal/runnable"
	"github.com/zclconf/go-cty/cty"
)

const (
	DataAttrValue = "value"
)

func (s *Data) Prepare(conductor *Conductor, skip bool, overridden bool) hcl.Diagnostics {
	return nil // no-op
}

func (s *Data) Run(conductor *Conductor, options ...runnable.Option) (diags hcl.Diagnostics) {
	logger := s.Logger()
	logger.Debugf("running %s.%s.%s", DataBlock, s.Provider, s.Id)
	hclContext := global.HclEvalContext()
	ctx := conductor.Context()

	var d hcl.Diagnostics

	cfg := runnable.NewConfig(options...)
	opts := []dataBlock.ProviderOption{
		dataBlock.WithPaths(cfg.Paths),
		dataBlock.WithBehavior(cfg.Behavior),
	}

	// region: mutating the data map
	// TODO: move it to a dedicated helper function

	// -> update r.Value accordingly
	var validProvider bool
	var value string
	var attr map[string]cty.Value
	for _, pr := range dataBlock.DefaultProviders {
		if pr.Name() == s.Provider {
			validProvider = true
			provide := pr.New()
			provide.SetContext(ctx)
			diags = diags.Extend(provide.DecodeBody(s.Body, opts...))
			value, d = provide.Value(ctx, s.Id, opts...)
			diags = diags.Extend(d)
			attr, d = provide.Attributes(ctx, s.Id, opts...)
			diags = diags.Extend(d)
			break
		}
	}
	if !validProvider {
		diags = diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("invalid provider %s", s.Provider),
			Detail:   fmt.Sprintf("built-in providers are %s", dataBlock.DefaultProviders),
		})
		return diags
	}

	if diags.HasErrors() {
		return diags
	}

	m := make(map[string]cty.Value)
	m[DataAttrValue] = cty.StringVal(value)
	for k, v := range attr {
		m[k] = v
	}

	global.DataBlockEvalContextMutex.Lock()

	global.EvalContextMutex.RLock()

	data := hclContext.Variables[DataBlock]

	var dataMutated map[string]cty.Value
	if data.IsNull() {
		dataMutated = make(map[string]cty.Value)
	} else {
		dataMutated = data.AsValueMap()
	}
	provider := dataMutated[s.Provider]
	var providerMutated map[string]cty.Value
	if provider.IsNull() {
		providerMutated = make(map[string]cty.Value)
	} else {
		providerMutated = provider.AsValueMap()
	}
	providerMutated[s.Id] = cty.ObjectVal(m)
	dataMutated[s.Provider] = cty.ObjectVal(providerMutated)
	global.EvalContextMutex.RUnlock()

	global.EvalContextMutex.Lock()
	hclContext.Variables[DataBlock] = cty.ObjectVal(dataMutated)
	global.EvalContextMutex.Unlock()

	global.DataBlockEvalContextMutex.Unlock()
	// endregion

	if diags.HasErrors() {
		return diags
	}

	v := s.Variables()
	logger.Debug(fmt.Sprintf("%s variables: %v", s.Identifier(), v))
	return nil
}

func (s *Data) CanRun(conductor *Conductor, options ...runnable.Option) (bool, hcl.Diagnostics) {
	return true, nil
}

func (s *Data) Terminated() bool {
	return true
}
