package orchestra

import (
	"context"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/srevinsaju/togomak/v1/pkg/c"
	"github.com/srevinsaju/togomak/v1/pkg/ci"
)

func ExpandImports(ctx context.Context, pipe *ci.Pipeline, parser *hclparse.Parser) (*ci.Pipeline, hcl.Diagnostics) {
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
		pwd := ctx.Value(c.TogomakContextCwd).(string)

		pipe, d = pipe.ExpandImports(ctx, parser, pwd)
		diags = diags.Extend(d)
		pipe.Logger().Debugf("expanding imports completed with %d error(s)", len(d.Errs()))

	}
	return pipe, diags
}
