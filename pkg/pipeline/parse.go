package pipeline

import (
	"context"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/srevinsaju/togomak/v1/pkg/c"
	"github.com/srevinsaju/togomak/v1/pkg/ci"
	"path/filepath"
)

func Read(ctx context.Context, parser *hclparse.Parser) (*ci.Pipeline, hcl.Diagnostics) {

	filePath := ctx.Value(c.TogomakContextPipelineFilePath).(string)
	if filePath == "" {
		filePath = "togomak.hcl"
	}
	owd := ctx.Value(c.TogomakContextOwd).(string)

	if filepath.IsAbs(filePath) == false {
		filePath = filepath.Join(owd, filePath)
	}
	f, diags := parser.ParseHCLFile(filePath)

	if diags.HasErrors() {
		return nil, diags
	}

	pipeline := &ci.Pipeline{}
	diags = gohcl.DecodeBody(f.Body, nil, pipeline)
	return pipeline, diags
}
