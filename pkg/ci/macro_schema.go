package ci

type macroSourceSpec struct {
	stages []string
}

type Macro struct {
	Id         string   `hcl:"id,label" json:"id"`
	Source     string   `hcl:"source,optional" json:"source"`
	Parameters []string `hcl:"parameters,optional" json:"parameters"`
	Stage      *Stage   `hcl:"stage,block" json:"stage"`

	sourceSpec *macroSourceSpec
}

type Macros []Macro
