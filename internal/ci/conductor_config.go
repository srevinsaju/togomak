package ci

import (
	"github.com/srevinsaju/togomak/v1/internal/behavior"
	"github.com/srevinsaju/togomak/v1/internal/path"
	"github.com/srevinsaju/togomak/v1/internal/rules"
)

type ConfigPipeline struct {
	Filtered    rules.Operations
	FilterQuery QueryEngines
	DryRun      bool
}

type Interface struct {
	// Verbosity is the level of verbosity
	Verbosity   int
	JSONLogging bool
}

type ConductorConfig struct {
	User     string
	Hostname string

	Paths *path.Path

	Interface Interface

	// Pipeline is the pipeline configuration
	Pipeline ConfigPipeline

	// Behavior is the behavior of the program
	Behavior *behavior.Behavior
}
