package ci

import "github.com/hashicorp/hcl/v2"

func (e *StageEnvironment) Variables() []hcl.Traversal {
	var traversal []hcl.Traversal
	traversal = append(traversal, e.Value.Variables()...)
	return traversal
}

func (s *Stage) Variables() []hcl.Traversal {
	var traversal []hcl.Traversal
	traversal = append(traversal, s.DependsOn.Variables()...)
	traversal = append(traversal, s.ForEach.Variables()...)
	traversal = append(traversal, s.Script.Variables()...)
	traversal = append(traversal, s.Args.Variables()...)

	// we leave out the "use" attribute since we will evaluate it separately

	for _, env := range s.Environment {
		traversal = append(traversal, env.Variables()...)
	}

	if s.Use != nil && s.Use.Parameters != nil {
		traversal = append(traversal, s.Use.Parameters.Variables()...)
	}
	return traversal
}

func (s Stages) Variables() []hcl.Traversal {
	var traversal []hcl.Traversal
	for _, stage := range s {
		traversal = append(traversal, stage.Variables()...)
	}
	return traversal
}
