package ci

import "github.com/hashicorp/hcl/v2"

func (s *Data) Override() bool {
	return false
}

// CheckIfDistinct checks if the data in d and dd are distinct
// TODO: check if this is a good way to do this
func (d Datas) CheckIfDistinct(dd Datas) hcl.Diagnostics {
	var diags hcl.Diagnostics
	for _, block := range d {
		for _, block2 := range dd {
			if block.Id == block2.Id {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Duplicate data block",
					Detail:   "Data with id " + block.Id + " is defined more than once",
				})
			}
		}
	}
	return diags
}
