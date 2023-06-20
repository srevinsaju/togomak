package ci

import (
	"context"
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/srevinsaju/togomak/v1/pkg/diag"
)

const MacroBlock = "macro"
const ParamBlock = "param"

func (m *Macro) Description() string {
	return ""
}

func (m *Macro) Identifier() string {
	return m.Id
}

func (m Macros) ById(id string) (*Macro, error) {
	for _, macro := range m {
		if macro.Id == id {
			return &macro, nil
		}
	}
	return nil, fmt.Errorf("macro block with id %s not found", id)
}

func (m *Macro) Type() string {
	return MacroBlock
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

func (m *Macro) Terminate() diag.Diagnostics {
	return nil
}

func (m *Macro) Kill() diag.Diagnostics {
	return nil
}

func (m *Macro) Expand(ctx context.Context) (Runnables, diag.Diagnostics) {
	return nil, nil
}
func (m *Macro) Expanded() bool {
	return false
}
func (m *Macro) ForEachDerived() bool {
	return false
}
