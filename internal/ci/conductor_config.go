package ci

import (
	"github.com/srevinsaju/togomak/v1/internal/behavior"
	"github.com/srevinsaju/togomak/v1/internal/path"
	"github.com/srevinsaju/togomak/v1/internal/rules"
)

// ConfigPipeline is the pipeline configuration
// provided at runtime, which includes filtered list of stages Filtered
// that need to run, and the query engines FilterQuery to filter them.
// It also includes DryRun which determines if the pipeline will not make
// actual shell executions
type ConfigPipeline struct {
	Filtered    rules.Operations
	FilterQuery QueryEngines
	DryRun      bool
}

// Interface is the interface configuration
// it includes the log output type configured using JSONLogging and the
// Verbosity level
type Interface struct {
	// Verbosity is the level of verbosity
	Verbosity   int
	JSONLogging bool
}

// ConductorConfig is the configuration provided at initialization
// of the conductor.
// Conductor may choose to update the values of the initial configuration
// at runtime, but the initial configuration is immutable.
type ConductorConfig struct {
	// User is the username of the user running the program
	User string

	// Hostname is the hostname of the machine
	Hostname string

	// Paths includes the path specifications such as current working directories,
	// original working directories, module path, and pipeline path
	Paths *path.Path

	// Interface is the interface configuration
	Interface Interface

	// Pipeline is the pipeline configuration
	Pipeline ConfigPipeline

	// Behavior is the behavior of the program
	Behavior *behavior.Behavior

	// Variables is the list of variables provided at runtime
	// and initialization using the -var flag, -var-file parsing or the environment
	// variables
	Variables Variables
}
