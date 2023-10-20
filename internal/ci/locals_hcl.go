package ci

import "github.com/hashicorp/hcl/v2"

func (l *Local) Variables() []hcl.Traversal {
	return l.Value.Variables()
}
