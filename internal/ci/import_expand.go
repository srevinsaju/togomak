package ci

import (
	"crypto/sha256"
	"fmt"
	"github.com/hashicorp/go-getter"
	"github.com/hashicorp/hcl/v2"
	"github.com/srevinsaju/togomak/v1/internal/ui"
	"path/filepath"
)

func (m *Import) Expand(conductor *Conductor, pwd string, dst string) (*Pipeline, hcl.Diagnostics) {
	logger := conductor.Logger().WithField("import", "")
	var diags hcl.Diagnostics
	shaIdentifier := sha256.Sum256([]byte(m.Identifier()))
	clientImportPath, err := filepath.Abs(filepath.Join(dst, fmt.Sprintf("%x", shaIdentifier)))
	if err != nil {
		panic(err)
	}

	// fmt.Println(pwd, dst, m.Source, fmt.Sprintf("%x", shaIdentifier))
	get := getter.Client{
		Ctx: conductor.Context(),
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

	p, d := ReadDirFromPath(conductor, clientImportPath)
	diags = diags.Extend(d)
	if diags.HasErrors() {
		return nil, diags
	}
	if p.Imports != nil {
		d := p.Imports.PopulateProperties(conductor)
		if d.HasErrors() {
			return nil, d
		}
		p, d = p.ExpandImports(conductor, clientImportPath)
		diags = diags.Extend(d)
	}
	return p, diags

}
