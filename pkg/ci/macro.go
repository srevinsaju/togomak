package ci

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
)

const MacroBlock = "macro"
const ParamBlock = "param"

func (m Macro) Description() string {
	return ""
}

func (m Macro) Identifier() string {
	return m.Id
}

func (m Macros) ById(id string) (*Macro, error) {
	for _, macro := range m {
		if macro.Id == id {
			return &macro, nil
		}
	}
	return nil, fmt.Errorf("data block with id %s not found", id)
}

func (m Macro) Type() string {
	return MacroBlock
}

func (m Macro) Variables() []hcl.Traversal {
	var traversal []hcl.Traversal
	return traversal
}

func (m Macro) IsDaemon() bool {
	return false
}
