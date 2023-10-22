package ci

import "github.com/hashicorp/hcl/v2"

func (m *Module) Variables() []hcl.Traversal {
	return m.Source.Variables()
}
