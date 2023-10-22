package ci

import (
	"github.com/hashicorp/hcl/v2"
)

type Module struct {
	Id        string         `hcl:"id,label" json:"id"`
	DependsOn hcl.Expression `hcl:"depends_on,optional" json:"depends_on"`
	Condition hcl.Expression `hcl:"if,optional" json:"if"`

	Source hcl.Expression `hcl:"source" json:"source"`

	pipeline *Pipeline

	Retry  *StageRetry  `hcl:"retry,block" json:"retry"`
	Daemon *StageDaemon `hcl:"daemon,block" json:"daemon"`
}

type Modules []Module
