package ci

import "github.com/hashicorp/hcl/v2"

type macroSourceSpec struct {
	stages []string
}

type Macro struct {
	Id string `hcl:"id,label" json:"id"`

	Source string         `hcl:"source,optional" json:"source"`
	Files  hcl.Expression `hcl:"files,optional" json:"files"`

	Parameters []string `hcl:"parameters,optional" json:"parameters"`
	Stage      *Stage   `hcl:"stage,block" json:"stage"`

	sourceSpec *macroSourceSpec
}

type Macros []Macro
