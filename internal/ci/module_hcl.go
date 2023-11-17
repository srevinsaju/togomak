package ci

import "github.com/hashicorp/hcl/v2"

func (m *Module) Variables() []hcl.Traversal {
	var vars []hcl.Traversal
	vars = append(vars, m.Source.Variables()...)
	vars = append(vars, m.DependsOn.Variables()...)
	vars = append(vars, m.Condition.Variables()...)
	vars = append(vars, m.ForEach.Variables()...)
	return vars
}
