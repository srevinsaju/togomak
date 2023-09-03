package ci

import "github.com/hashicorp/hcl/v2"

func (m *Macro) Override() bool {
	return false
}

// CheckIfDistinct checks if the macro in m and mm are distinct
func (m Macros) CheckIfDistinct(mm Macros) hcl.Diagnostics {
	for _, macro := range m {
		for _, macro2 := range mm {
			if macro.Identifier() == macro2.Identifier() {
				return hcl.Diagnostics{
					{
						Severity: hcl.DiagError,
						Summary:  "Duplicate macro",
						Detail:   "Macro with id " + macro.Id + " is defined more than once",
					},
				}
			}
		}
	}
	return nil
}
