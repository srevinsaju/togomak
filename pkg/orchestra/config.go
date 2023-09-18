package orchestra

import (
	"github.com/srevinsaju/togomak/v1/pkg/filter"
)

type ConfigPipeline struct {
	FilePath string
	Filtered filter.FilterList
	DryRun   bool
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

	Verbosity int

	Pipeline ConfigPipeline
}
