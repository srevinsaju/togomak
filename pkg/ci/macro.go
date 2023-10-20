package ci

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/srevinsaju/togomak/v1/pkg/blocks"
)

func (m *Macro) Description() string {
	return ""
}

func (m *Macro) Identifier() string {
	return m.Id
}

func (m Macros) ById(id string) (*Macro, hcl.Diagnostics) {
	for _, macro := range m {
		if macro.Id == id {
			return &macro, nil
		}
	}
	return nil, hcl.Diagnostics{
		{
			Severity: hcl.DiagError,
			Summary:  "Macro not found",
			Detail:   "Macro with id " + id + " not found",
		},
	}
}

func (m *Macro) Type() string {
	return blocks.MacroBlock
}

func (m *Macro) Variables() []hcl.Traversal {
	var traversal []hcl.Traversal

	traversal = append(traversal, m.Files.Variables()...)
	if m.Stage != nil {
		traversal = append(traversal, m.Stage.Variables()...)

	}
	return traversal
}

func (m *Macro) IsDaemon() bool {
	return false
}

func (m *Macro) Terminate(conductor *Conductor, safe bool) hcl.Diagnostics {
	return nil
}

func (m *Macro) Kill() hcl.Diagnostics {
	return nil
}

func (m *Macro) Set(k any, v any) {
}

func (m *Macro) Get(k any) any {
	return nil
}
