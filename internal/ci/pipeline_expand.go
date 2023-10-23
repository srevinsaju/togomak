package ci

import (
	"github.com/hashicorp/hcl/v2"
	"path/filepath"
)

func (pipe *Pipeline) ExpandImports(conductor *Conductor, pwd string) (*Pipeline, hcl.Diagnostics) {
	var pipes MetaList
	var diags hcl.Diagnostics
	pipes = pipes.Append(NewMeta(pipe, nil, "memory"))
	tmpDir := conductor.TempDir()

	dst, err := filepath.Abs(filepath.Join(tmpDir, "import"))
	if err != nil {
		panic(err)
	}
	m := pipe.Imports
	for _, im := range m {
		p, d := im.Expand(conductor, pwd, dst)
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
