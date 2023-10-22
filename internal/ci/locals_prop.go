package ci

import "github.com/hashicorp/hcl/v2"

func (l *Local) Override() bool {
	return false
}

// CheckIfDistinct checks if the locals in l and ll are distinct
func (l LocalGroup) CheckIfDistinct(ll LocalGroup) hcl.Diagnostics {
	var diags hcl.Diagnostics
	for _, local := range l {
		for _, local2 := range ll {
			if local.Identifier() == local2.Identifier() {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Duplicate local",
					Detail:   "Local with id " + local.Identifier() + " is defined more than once",
				})
			}
		}
	}
	return diags
}
