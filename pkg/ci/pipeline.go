package ci

import "github.com/hashicorp/hcl/v2"

const PipelineBlock = "pipeline"

type Pipeline struct {
	Builder Builder `hcl:"togomak,block" json:"togomak"`

	Stages        Stages        `hcl:"stage,block" json:"stages"`
	Data          Datas         `hcl:"data,block" json:"data"`
	Macros        Macros        `hcl:"macro,block" json:"macro"`
	DataProviders DataProviders `hcl:"provider,block" json:"providers"`
}

func (p Pipeline) Variables() []hcl.Traversal {
	var traversal []hcl.Traversal
	traversal = append(traversal, p.Stages.Variables()...)
	traversal = append(traversal, p.Data.Variables()...)
	traversal = append(traversal, p.DataProviders.Variables()...)
	return traversal
}
