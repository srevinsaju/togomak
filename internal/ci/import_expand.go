package ci

import (
	"context"
	"crypto/sha256"
	"fmt"
	"github.com/hashicorp/go-getter"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/srevinsaju/togomak/v1/internal/global"
	"github.com/srevinsaju/togomak/v1/internal/ui"
	"path/filepath"
)

func (m *Import) Expand(ctx context.Context, parser *hclparse.Parser, pwd string, dst string) (*Pipeline, hcl.Diagnostics) {
	logger := global.Logger().WithField("import", "")
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
		p, d = p.ExpandImports(ctx, parser, clientImportPath)
		diags = diags.Extend(d)
	}
	return p, diags

}
