package conductor

import (
	"github.com/srevinsaju/togomak/v1/pkg/behavior"
	"github.com/srevinsaju/togomak/v1/pkg/path"
	"github.com/srevinsaju/togomak/v1/pkg/rules"
)

type ConfigPipeline struct {
	Filtered    rules.Operations
	FilterQuery rules.QueryEngines
	DryRun      bool
}

type Interface struct {
	// Verbosity is the level of verbosity
	Verbosity int
}

type Config struct {
	User     string
	Hostname string

	Paths *path.Path

	Interface Interface

	// Pipeline is the pipeline configuration
	Pipeline ConfigPipeline

	// Behavior is the behavior of the program
	Behavior *behavior.Behavior
}
