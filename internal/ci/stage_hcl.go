package ci

import (
	"github.com/hashicorp/hcl/v2"
)

func (e *StageEnvironment) Variables() []hcl.Traversal {
	var traversal []hcl.Traversal
	traversal = append(traversal, e.Value.Variables()...)
	return traversal
}
func (e *StageContainerVolume) Variables() []hcl.Traversal {
	var traversal []hcl.Traversal
	traversal = append(traversal, e.Source.Variables()...)
	traversal = append(traversal, e.Destination.Variables()...)
	return traversal
}

func (e *StageDaemon) Variables() []hcl.Traversal {
	var traversal []hcl.Traversal
	if e.Lifecycle != nil {
		traversal = append(traversal, e.Lifecycle.Variables()...)
	}
	return traversal
}

func (e *StageContainerVolumes) Variables() []hcl.Traversal {
	var traversal []hcl.Traversal
	for _, volume := range *e {
		traversal = append(traversal, volume.Variables()...)
	}
	return traversal
}

func (s *CoreStage) Variables() []hcl.Traversal {
	var traversal []hcl.Traversal
	traversal = append(traversal, s.Condition.Variables()...)
	traversal = append(traversal, s.Dir.Variables()...)
	traversal = append(traversal, s.DependsOn.Variables()...)
	traversal = append(traversal, s.Script.Variables()...)
	traversal = append(traversal, s.Args.Variables()...)

	traversal = append(traversal, s.dependsOnVariablesMacro...)

	if s.Use != nil {
		traversal = append(traversal, s.Use.Macro.Variables()...)
		traversal = append(traversal, s.Use.Parameters.Variables()...)
	}
	if s.Container != nil {
		traversal = append(traversal, s.Container.Volumes.Variables()...)
	}
	if s.Daemon != nil {
		traversal = append(traversal, s.Daemon.Variables()...)
	}

	for _, env := range s.Environment {
		traversal = append(traversal, env.Variables()...)
	}
	if s.PostHook != nil {
		for _, hook := range s.PostHook {
			traversal = append(traversal, hook.Stage.Variables()...)
		}
	}
	if s.PreHook != nil {
		for _, hook := range s.PreHook {
			traversal = append(traversal, hook.Stage.Variables()...)
		}
	}
	return traversal
}

func (s *Stage) Variables() []hcl.Traversal {
	var traversal []hcl.Traversal
	traversal = append(traversal, s.CoreStage.Variables()...)
	if s.Lifecycle != nil {
		traversal = append(traversal, s.Lifecycle.Timeout.Variables()...)
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
