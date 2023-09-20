package orchestra

import (
	"context"
	"github.com/hashicorp/hcl/v2"
	"github.com/srevinsaju/togomak/v1/pkg/ci"
	"github.com/srevinsaju/togomak/v1/pkg/pipeline"
)

func ExpandImports(pipe *ci.Pipeline, ctx context.Context, t Togomak) (*ci.Pipeline, hcl.Diagnostics) {
	var d hcl.Diagnostics
	var diags hcl.Diagnostics

	if len(pipe.Imports) != 0 {
		t.Logger.Debugf("expanding imports")
		d = pipe.Imports.PopulateProperties()
		diags = diags.Extend(d)
		if d.HasErrors() {
			return pipe, diags
		}
		t.Logger.Debugf("populating properties for imports completed with %d error(s)", len(d.Errs()))
		pipe, d = pipeline.ExpandImports(ctx, pipe, t.parser)
		diags = diags.Extend(d)
		t.Logger.Debugf("expanding imports completed with %d error(s)", len(d.Errs()))

	}
	return pipe, diags
}
