package ci

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/srevinsaju/togomak/v1/internal/blocks"
)

func (v *Variable) Name() string {
	return v.Id
}

func (v *Variable) Description() Description {
	return Description{
		Name:        v.Id,
		Description: v.Desc,
	}
}

func (v *Variable) Identifier() string {
	return v.Id
}

func (v *Variable) Set(k any, value any) {
	return // do nothing
}

func (v *Variable) Get(k any) any {
	return nil
}

func (v *Variable) Type() string {
	return blocks.VariableBlock
}

func (v *Variable) IsDaemon() bool {
	return false
}

func (s Variables) ById(id string) (*Variable, hcl.Diagnostics) {
	for _, variable := range s {
		if variable.Id == id {
			return variable, nil
		}
	}
	return nil, hcl.Diagnostics{
		{
			Severity: hcl.DiagError,
			Summary:  "Variable not found",
			Detail:   fmt.Sprintf("variable input with id %s not found", id),
		},
	}
}

// CheckIfDistinct checks if the stages in s and ss are distinct
// TODO: check if this is a good way to do this
func (s Variables) CheckIfDistinct(ss Variables) hcl.Diagnostics {
	var diags hcl.Diagnostics
	for _, stage := range s {
		for _, stage2 := range ss {
			if stage.Id == stage2.Id {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Duplicate variable",
					Detail:   "Stage with id " + stage.Id + " is defined more than once",
				})
			}
		}
	}
	return diags
}
