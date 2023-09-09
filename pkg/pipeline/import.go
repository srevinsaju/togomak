package pipeline

import (
	"context"
	"crypto/sha256"
	"fmt"
	"github.com/hashicorp/go-getter"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/srevinsaju/togomak/v1/pkg/c"
	"github.com/srevinsaju/togomak/v1/pkg/ci"
	"path/filepath"
)

func expandImport(m ci.Import, ctx context.Context, pwd string, dst string) (*ci.Pipeline, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	shaIdentifier := sha256.Sum256([]byte(m.Source))
	clientImportPath, err := filepath.Abs(filepath.Join(dst, fmt.Sprintf("%x", shaIdentifier)))
	if err != nil {
		panic(err)
	}

	// fmt.Println(pwd, dst, m.Source, fmt.Sprintf("%x", shaIdentifier))
	get := getter.Client{
		Ctx: ctx,
		Src: m.Source,
		Dir: true,
		Pwd: pwd,
		// TODO: implement progress tracker
		Dst: clientImportPath,
	}
	err = get.Get()
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
	if diags.HasErrors() {
		return nil, diags
	}
	if p.Imports != nil {
		p, d = expandImports(ctx, p, clientImportPath)
		diags = diags.Extend(d)
	}
	return p, diags

}

func ExpandImports(ctx context.Context, pipe *ci.Pipeline) (*ci.Pipeline, hcl.Diagnostics) {
	pwd := ctx.Value(c.TogomakContextCwd).(string)
	return expandImports(ctx, pipe, pwd)
}

func expandImports(ctx context.Context, pipe *ci.Pipeline, pwd string) (*ci.Pipeline, hcl.Diagnostics) {
	var pipes MetaList
	var diags hcl.Diagnostics
	pipes = pipes.Append(NewMeta(pipe, nil, "memory"))
	tmpDir := ctx.Value(c.TogomakContextTempDir).(string)

	dst, err := filepath.Abs(filepath.Join(tmpDir, "import"))
	if err != nil {
		panic(err)
	}
	m := pipe.Imports
	for _, im := range m {
		p, d := expandImport(im, ctx, pwd, dst)
		diags = diags.Extend(d)
		pipes = pipes.Append(NewMeta(p, nil, im.Source))
	}
	p, d := Merge(pipes)
	diags = diags.Extend(d)
	return p, diags
}
