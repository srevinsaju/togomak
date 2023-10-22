package ci

import (
	"context"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/srevinsaju/togomak/v1/internal/global"
	"path/filepath"
)

func (pipe *Pipeline) ExpandImports(ctx context.Context, parser *hclparse.Parser, pwd string) (*Pipeline, hcl.Diagnostics) {
	var pipes MetaList
	var diags hcl.Diagnostics
	pipes = pipes.Append(NewMeta(pipe, nil, "memory"))
	tmpDir := global.TempDir()

	dst, err := filepath.Abs(filepath.Join(tmpDir, "import"))
	if err != nil {
		panic(err)
	}
	m := pipe.Imports
	for _, im := range m {
		p, d := im.Expand(ctx, parser, pwd, dst)
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
