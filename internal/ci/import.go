package ci

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/srevinsaju/togomak/v1/internal/global"
	"github.com/zclconf/go-cty/cty"
)

const ImportBlock = "import"

func (m *Import) Description() string {
	return ""
}

func (m *Import) Identifier() string {
	if m.id == "" {
		panic("id not set")
	}
	return m.id
}

func (i Imports) ById(id string) (*Import, hcl.Diagnostics) {
	for _, macro := range i {
		if macro.Identifier() == id {
			return macro, nil
		}
	}
	return nil, hcl.Diagnostics{
		{
			Severity: hcl.DiagError,
			Summary:  "import not found",
			Detail:   "import with id " + id + " not found",
		},
	}
}

func (i Imports) PopulateProperties() hcl.Diagnostics {
	var diags hcl.Diagnostics
	for _, imp := range i {
		diags = diags.Extend(imp.populateProperties())
	}
	return diags
}

func (m *Import) Type() string {
	return ImportBlock
}

func (m *Import) Variables() []hcl.Traversal {
	return nil
}

func (m *Import) IsDaemon() bool {
	return false
}

func (m *Import) Terminate(safe bool) hcl.Diagnostics {
	return nil
}

func (m *Import) Kill() hcl.Diagnostics {
	return nil
}

func (m *Import) Set(k any, v any) {
}

func (m *Import) Get(k any) any {
	return nil
}

func (m *Import) populateProperties() hcl.Diagnostics {
	evalContext := global.HclEvalContext()
	s, diags := m.Source.Value(evalContext)

	if s.Type() != cty.String {
		diags = diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "invalid source",
			Detail:   "source should be a string",
		})
		return diags
	}

	if diags.HasErrors() {
		return diags
	}
	m.id = s.AsString()
	return diags
}
