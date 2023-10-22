package ci

import (
	"github.com/hashicorp/hcl/v2"
)

type Module struct {
	Id   string `hcl:"id,label" json:"id"`
	Name string `hcl:"name,optional" json:"name"`

	DependsOn hcl.Expression `hcl:"depends_on,optional" json:"depends_on"`
	Condition hcl.Expression `hcl:"if,optional" json:"if"`
	ForEach   hcl.Expression `hcl:"for_each,optional" json:"for_each"`

	Source hcl.Expression `hcl:"source" json:"source"`

	pipeline *Pipeline

	Lifecycle *Lifecycle   `hcl:"lifecycle,block" json:"lifecycle"`
	Retry     *StageRetry  `hcl:"retry,block" json:"retry"`
	Daemon    *StageDaemon `hcl:"daemon,block" json:"daemon"`

	Body hcl.Body `hcl:",remain" json:"body"`
}

type Modules []Module
