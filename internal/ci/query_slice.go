package ci

import (
	"github.com/hashicorp/hcl/v2"
)

type QueryEngines []*QueryEngine

func NewSlice(queries []string) (QueryEngines, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	var engines QueryEngines
	for _, query := range queries {
		e, d := New(query)
		diags = diags.Extend(d)
		engines = append(engines, e)
	}
	return engines, diags
}

func (e QueryEngines) Eval(ok bool, stage PhasedBlock) (bool, bool, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	var d hcl.Diagnostics

	var overridden bool
	var resultOk bool

	for _, engine := range e {
		resultOk, overridden, d = engine.Eval(ok, stage)
		diags = diags.Extend(d)
		if d.HasErrors() {
			continue
		}
		if resultOk {
			return resultOk, overridden, diags
		}
	}
	return resultOk, overridden, diags
}
