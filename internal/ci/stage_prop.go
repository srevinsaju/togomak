package ci

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/srevinsaju/togomak/v1/internal/blocks"
	"github.com/srevinsaju/togomak/v1/internal/x"
)

func (s *Stage) Override() bool {
	return false
}

func (s Stages) Override() bool {
	return false
}

// CheckIfDistinct checks if the stages in s and ss are distinct
// TODO: check if this is a good way to do this
func (s Stages) CheckIfDistinct(ss Stages) hcl.Diagnostics {
	var diags hcl.Diagnostics
	for _, stage := range s {
		for _, stage2 := range ss {
			if stage.Id == stage2.Id {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Duplicate stage",
					Detail:   "Stage with id " + stage.Id + " is defined more than once",
				})
			}
		}
	}
	return diags
}

func (s *Stage) String() string {
	return x.RenderBlock(blocks.StageBlock, s.Id)
}
