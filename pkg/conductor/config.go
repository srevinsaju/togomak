package conductor

import (
	"github.com/srevinsaju/togomak/v1/pkg/behavior"
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

type Config struct {
	User     string
	Hostname string

	Paths path.Path

	Interface Interface

	// Pipeline is the pipeline configuration
	Pipeline ConfigPipeline

	// Behavior is the behavior of the program
	Behavior behavior.Behavior
}
