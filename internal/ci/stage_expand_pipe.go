package ci

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/sirupsen/logrus"
	"github.com/srevinsaju/togomak/v1/internal/conductor"
	"github.com/srevinsaju/togomak/v1/internal/runnable"
	"github.com/zclconf/go-cty/cty"
	"os/exec"
	"sync"
)

type stageEval struct {
	rootEvalContext *hcl.EvalContext
	evalContext     *hcl.EvalContext
	mu              *sync.RWMutex
}

func (e *stageEval) Mutex() *sync.RWMutex {
	return e.mu
}

func (e *stageEval) Context() *hcl.EvalContext {
	return e.evalContext
}

type stageAttr struct {
	environment        map[string]cty.Value
	environmentStrings []string

	cmd *exec.Cmd
}

type stageConductor struct {
	conductor *Conductor
	params    map[string]cty.Value

	attr *stageAttr

	eval *stageEval

	cfg *runnable.Config

	diags hcl.Diagnostics
	*Stage
}

func (s *Stage) newStageConductor(conductor *Conductor, cfg *runnable.Config) *stageConductor {
	return &stageConductor{
		conductor: conductor,
		params:    map[string]cty.Value{},
		eval: &stageEval{
			rootEvalContext: conductor.Eval().Context(),
			evalContext:     conductor.Eval().Context(),
			mu:              conductor.Eval().Mutex(),
		},
		cfg:   cfg,
		diags: hcl.Diagnostics{},
		Stage: s,
	}
}

func (s *stageConductor) Logger() *logrus.Entry {
	return s.conductor.Logger().WithField("stage", s.Id)
}

func (s *stageConductor) Eval() conductor.Eval {
	return s.eval
}

type stagePipe func(ctx *stageConductor) *stageConductor
