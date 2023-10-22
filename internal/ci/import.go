package ci

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

const ImportBlock = "import"

func (m *Import) Description() Description {
	return Description{Name: m.id}
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

func (i Imports) PopulateProperties(conductor *Conductor) hcl.Diagnostics {
	var diags hcl.Diagnostics
	for _, imp := range i {
		diags = diags.Extend(imp.populateProperties(conductor))
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

func (m *Import) populateProperties(conductor *Conductor) hcl.Diagnostics {
	evalContext := conductor.Eval().Context()
	conductor.Eval().Mutex().RLock()
	s, diags := m.Source.Value(evalContext)
	conductor.Eval().Mutex().RUnlock()

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
