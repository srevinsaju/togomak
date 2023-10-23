package ci

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/srevinsaju/togomak/v1/internal/path"
)

func ExpandImports(conductor *Conductor, pipe *Pipeline, paths *path.Path) (*Pipeline, hcl.Diagnostics) {
	var d hcl.Diagnostics
	var diags hcl.Diagnostics

	if len(pipe.Imports) != 0 {
		pipe.Logger().Debugf("expanding imports")
		d = pipe.Imports.PopulateProperties(conductor)
		diags = diags.Extend(d)
		if d.HasErrors() {
			return pipe, diags
		}

		pipe.Logger().Debugf("populating properties for imports completed with %d error(s)", len(d.Errs()))

		pipe, d = pipe.ExpandImports(conductor, paths.Cwd)
		diags = diags.Extend(d)
		pipe.Logger().Debugf("expanding imports completed with %d error(s)", len(d.Errs()))

	}
	return pipe, diags
}
