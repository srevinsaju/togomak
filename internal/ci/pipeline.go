package ci

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/sirupsen/logrus"
	"github.com/srevinsaju/togomak/v1/internal/meta"
)

const PipelineBlock = "pipeline"

type Pipeline struct {
	Builder Builder `hcl:"togomak,block" json:"togomak"`

	Stages  Stages      `hcl:"stage,block" json:"stages"`
	Data    Datas       `hcl:"data,block" json:"data"`
	Macros  Macros      `hcl:"macro,block" json:"macro"`
	Locals  LocalsGroup `hcl:"locals,block" json:"locals"`
	Imports Imports     `hcl:"import,block" json:"import"`

	Modules Modules `hcl:"module,block" json:"modules"`

	DataProviders DataProviders `hcl:"provider,block" json:"providers"`

	// private stuff
	Local LocalGroup

	Pre  *PreStage  `hcl:"pre,block" json:"pre"`
	Post *PostStage `hcl:"post,block" json:"post"`
}

func (pipe *Pipeline) Variables() []hcl.Traversal {
	var traversal []hcl.Traversal
	traversal = append(traversal, pipe.Stages.Variables()...)
	traversal = append(traversal, pipe.Data.Variables()...)
	traversal = append(traversal, pipe.DataProviders.Variables()...)
	return traversal
}

func (pipe *Pipeline) Logger() *logrus.Entry {
	return logrus.WithField("pipeline", "")
}

func (pipe *Pipeline) Resolve(runnableId string) (Block, bool, hcl.Diagnostics) {
	var runnable Block
	var diags hcl.Diagnostics
	var d hcl.Diagnostics

	skip := false
	switch runnableId {
	case meta.RootStage:
		skip = true
	case meta.PreStage:
		if pipe.Pre == nil {
			pipe.Logger().Debugf("skipping runnable pre block %s, not defined", runnableId)
			skip = true
			break
		}
		runnable = pipe.Pre.ToStage()
	case meta.PostStage:
		if pipe.Post == nil {
			pipe.Logger().Debugf("skipping runnable post block %s, not defined", runnableId)
			skip = true
			break
		}
		runnable = pipe.Post.ToStage()
	default:
		runnable, d = Resolve(pipe, runnableId)
		diags = diags.Extend(d)
	}
	return runnable, skip, diags
}
