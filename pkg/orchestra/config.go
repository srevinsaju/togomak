package orchestra

import (
	"github.com/srevinsaju/togomak/v1/pkg/filter"
)

type ConfigPipeline struct {
	FilePath string
	Filtered filter.FilterList
	DryRun   bool
}

type Interface struct {
	// Verbosity is the level of verbosity
	Verbosity int
}

type Config struct {
	Owd string
	Dir string

	Unattended   bool
	Ci           bool
	Child        bool
	Parent       string
	ParentParams []string

	User     string
	Hostname string

	Interface Interface

	Pipeline ConfigPipeline
}
