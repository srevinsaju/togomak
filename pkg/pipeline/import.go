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
	"github.com/srevinsaju/togomak/v1/pkg/global"
	"github.com/srevinsaju/togomak/v1/pkg/ui"
	"path/filepath"
)

var logger = global.Logger().WithField("import", "")

func expandImport(m *ci.Import, ctx context.Context, parser *hclparse.Parser, pwd string, dst string) (*ci.Pipeline, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	shaIdentifier := sha256.Sum256([]byte(m.Identifier()))
	clientImportPath, err := filepath.Abs(filepath.Join(dst, fmt.Sprintf("%x", shaIdentifier)))
	if err != nil {
		panic(err)
	}

	// fmt.Println(pwd, dst, m.Source, fmt.Sprintf("%x", shaIdentifier))
	get := getter.Client{
		Ctx: ctx,
		Src: m.Identifier(),
		Dir: true,
		Pwd: pwd,
		Dst: clientImportPath,
	}
	ppb := ui.NewPassiveProgressBar(logger, fmt.Sprintf("pulling %s", m.Identifier()))
	ppb.Init()
	err = get.Get()
	if err != nil {
		diags = diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "import failed",
			Detail:   fmt.Sprintf("import of %s failed: %s", m.Identifier(), err.Error()),
		})
	}
	ppb.Done()

	p, d := ReadDirFromPath(clientImportPath, parser)
	diags = diags.Extend(d)
	if diags.HasErrors() {
		return nil, diags
	}
	if p.Imports != nil {
		d := p.Imports.PopulateProperties()
		if d.HasErrors() {
			return nil, d
		}
		p, d = expandImports(ctx, p, parser, clientImportPath)
		diags = diags.Extend(d)
	}
	return p, diags

}

func ExpandImports(ctx context.Context, pipe *ci.Pipeline, parser *hclparse.Parser) (*ci.Pipeline, hcl.Diagnostics) {
	pwd := ctx.Value(c.TogomakContextCwd).(string)
	return expandImports(ctx, pipe, parser, pwd)
}

func expandImports(ctx context.Context, pipe *ci.Pipeline, parser *hclparse.Parser, pwd string) (*ci.Pipeline, hcl.Diagnostics) {
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
		p, d := expandImport(im, ctx, parser, pwd, dst)
		diags = diags.Extend(d)
		if d.HasErrors() {
			continue
		}
		pipes = pipes.Append(NewMeta(p, nil, im.Identifier()))
	}
	p, d := Merge(pipes)
	diags = diags.Extend(d)
	return p, diags
}
