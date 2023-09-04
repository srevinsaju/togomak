package pipeline

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-getter"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/srevinsaju/togomak/v1/pkg/c"
	"github.com/srevinsaju/togomak/v1/pkg/ci"
	"github.com/srevinsaju/togomak/v1/pkg/x"
	"os"
	"path/filepath"
)

func ExpandImport(m ci.Import, ctx context.Context) (*ci.Pipeline, hcl.Diagnostics) {

	var diags hcl.Diagnostics
	tmpDir := ctx.Value(c.TogomakContextTempDir).(string)
	clientImportPath := filepath.Join(tmpDir, "import", m.Identifier())
	x.Must(os.MkdirAll(clientImportPath, 0755))

	get := getter.Client{
		Ctx: ctx,
		Src: m.Source,
		Dir: true,
		// TODO: implement progress tracker
		Dst: clientImportPath,
	}
	err := get.Get()
	if err != nil {
		diags = diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "import failed",
			Detail:   fmt.Sprintf("import of %s failed: %s", m.Source, err.Error()),
		})
	}
	parser := hclparse.NewParser()
	p, d := ReadDirFromPath(clientImportPath, parser)
	diags = diags.Extend(d)
	return p, diags

}

func ExpandImports(ctx context.Context, pipe *ci.Pipeline) (*ci.Pipeline, hcl.Diagnostics) {
	var pipes MetaList
	var diags hcl.Diagnostics
	pipes = pipes.Append(NewMeta(pipe, nil, "memory"))

	m := pipe.Imports
	for _, im := range m {
		p, d := ExpandImport(im, ctx)
		diags = diags.Extend(d)
		pipes = pipes.Append(NewMeta(p, nil, im.Source))
	}
	p, d := Merge(pipes)
	diags = diags.Extend(d)
	return p, diags
}
