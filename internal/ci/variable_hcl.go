package ci

import "github.com/hashicorp/hcl/v2"

func (v *Variable) Variables() []hcl.Traversal {
	var traversal []hcl.Traversal
	traversal = append(traversal, v.Value.Variables()...)
	return traversal

}
