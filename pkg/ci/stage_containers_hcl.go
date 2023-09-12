package ci

import (
	"github.com/docker/go-connections/nat"
	"github.com/hashicorp/hcl/v2"
	"github.com/srevinsaju/togomak/v1/pkg/global"
	"github.com/zclconf/go-cty/cty"
)

// Nat returns a map of exposed ports and a map of port bindings
// after parsing the HCL configuration from StageContainerPorts.
func (s StageContainerPorts) Nat(evalCtx *hcl.EvalContext) (map[nat.Port]struct{}, map[nat.Port][]nat.PortBinding, hcl.Diagnostics) {
	var hclDiags hcl.Diagnostics
	var rawPortSpecs []string
	for _, port := range s {
		global.EvalContextMutex.RLock()
		p, d := port.Port.Value(evalCtx)
		global.EvalContextMutex.RUnlock()

		hclDiags = hclDiags.Extend(d)
		if d.HasErrors() {
			continue
		}
		if p.Type() != cty.String {
			hclDiags = hclDiags.Append(&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Invalid port specification",
				Detail:   "Port specification must be a string",
				Subject:  port.Port.Range().Ptr(),
			})
			continue
		}

		rawPortSpecs = append(rawPortSpecs, p.AsString())
	}

	exposedPorts, bindings, err := nat.ParsePortSpecs(rawPortSpecs)
	if err != nil {
		hclDiags = hclDiags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid port specification",
			Detail:   err.Error(),
		})
	}
	return exposedPorts, bindings, hclDiags
}
