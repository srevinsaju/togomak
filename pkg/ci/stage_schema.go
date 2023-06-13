package ci

import "github.com/hashicorp/hcl/v2"

const StageBlock = "stage"

type StageContainer struct {
	Image   string   `hcl:"image" json:"image"`
	Volumes []string `hcl:"volumes" json:"volumes"`
}

type Stages []Stage

type StageEnvironment struct {
	Name  string         `hcl:"name" json:"name"`
	Value hcl.Expression `hcl:"value" json:"value"`
}

type StageRetry struct {
	Enabled            bool `hcl:"enabled" json:"enabled"`
	Attempts           int  `hcl:"attempts" json:"attempts"`
	ExponentialBackoff bool `hcl:"exponential_backoff" json:"exponential_backoff"`
	MinBackoff         int  `hcl:"min_backoff" json:"min_backoff"`
	MaxBackoff         int  `hcl:"max_backoff" json:"max_backoff"`
}

type StageUse struct {
	Macro      hcl.Expression `hcl:"macro" json:"macro"`
	Parameters hcl.Expression `hcl:"parameters,optional" json:"parameters"`
}

type StageDaemon struct {
	Enabled bool `hcl:"enabled" json:"enabled"`
}

type Stage struct {
	Id        string         `hcl:"id,label" json:"id"`
	Condition hcl.Expression `hcl:"if,optional" json:"if"`
	DependsOn hcl.Expression `hcl:"depends_on,optional" json:"depends_on"`
	ForEach   hcl.Expression `hcl:"for_each,optional" json:"for_each"`
	Use       *StageUse      `hcl:"use,block" json:"use"`

	Daemon *StageDaemon `hcl:"daemon,block" json:"daemon"`

	Name        string              `hcl:"name,optional" json:"name"`
	Dir         hcl.Expression      `hcl:"dir,optional" json:"dir"`
	Script      hcl.Expression      `hcl:"script,optional" json:"script"`
	Args        hcl.Expression      `hcl:"args,optional" json:"args"`
	Container   *StageContainer     `hcl:"container,block" json:"container"`
	Environment []*StageEnvironment `hcl:"env,block" json:"environment"`
	Retry       *StageRetry         `hcl:"retry,block" json:"retry"`
}
