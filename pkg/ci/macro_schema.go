package ci

type Macro struct {
	Id         string   `hcl:"id,label" json:"id"`
	Parameters []string `hcl:"parameters,optional" json:"parameters"`
	Stage      Stage    `hcl:"stage,block" json:"stage"`
}

type Macros []Macro
