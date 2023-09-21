package conductor

import (
	"github.com/srevinsaju/togomak/v1/pkg/filter"
	"github.com/srevinsaju/togomak/v1/pkg/path"
)

type ConfigPipeline struct {
	Filtered filter.FilterList
	DryRun   bool
}

type Interface struct {
	// Verbosity is the level of verbosity
	Verbosity int
}

type BehaviorChild struct {
	// Enabled is the flag to indicate whether the program is running in child mode
	Enabled bool

	// Parent is the flag to indicate whether the program is running in parent mode
	Parent string

	// ParentParams is the list of parameters to be passed to the parent
	ParentParams []string
}

type Behavior struct {
	// Unattended is the flag to indicate whether the program is running in unattended mode
	Unattended bool

	// Ci is the flag to indicate whether the program is running in CI mode
	Ci bool

	// Child is the flag to indicate whether the program is running in child mode
	Child BehaviorChild
}

type Config struct {
	User     string
	Hostname string

	Paths path.Path

	Interface Interface

	// Pipeline is the pipeline configuration
	Pipeline ConfigPipeline

	// Behavior is the behavior of the program
	Behavior Behavior
}
