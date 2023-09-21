package orchestra

import (
	"context"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/srevinsaju/togomak/v1/pkg/ci"
	"github.com/srevinsaju/togomak/v1/pkg/path"
)

func ExpandImports(ctx context.Context, pipe *ci.Pipeline, parser *hclparse.Parser, paths *path.Path) (*ci.Pipeline, hcl.Diagnostics) {
	var d hcl.Diagnostics
	var diags hcl.Diagnostics

	if len(pipe.Imports) != 0 {
		pipe.Logger().Debugf("expanding imports")
		d = pipe.Imports.PopulateProperties()
		diags = diags.Extend(d)
		if d.HasErrors() {
			return pipe, diags
		}

		pipe.Logger().Debugf("populating properties for imports completed with %d error(s)", len(d.Errs()))

		pipe, d = pipe.ExpandImports(ctx, parser, paths.Cwd)
		diags = diags.Extend(d)
		pipe.Logger().Debugf("expanding imports completed with %d error(s)", len(d.Errs()))

	}
	return pipe, diags
}
