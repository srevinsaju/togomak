package ci

import (
	"github.com/hashicorp/hcl/v2"
)

const ImportBlock = "import"

func (m *Import) Description() string {
	return ""
}

func (m *Import) Identifier() string {
	return m.Source
}

func (m Imports) ById(id string) (*Import, hcl.Diagnostics) {
	for _, macro := range m {
		if macro.Source == id {
			return &macro, nil
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
