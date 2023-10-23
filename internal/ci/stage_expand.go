package ci

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/srevinsaju/togomak/v1/internal/blocks"
	"github.com/zclconf/go-cty/cty"
)

// expandParams expands the parameters in the stage, if any.
// this includes globally passed param. blocks, or from macros including the parameter attributes
func (s *stageConductor) expandParams() *stageConductor {
	// reads the parameters from the caller - global parameters
	// mostly kept for backward compatibility for macros
	logger := s.Logger().WithField("stage", s.Id)
	logger.Debugf("expanding macro parameters")
	params := map[string]cty.Value{}
	var diags hcl.Diagnostics

	s.Eval().Mutex().RLock()
	oldParam, ok := s.conductor.Eval().Context().Variables[blocks.ParamBlock]
	s.Eval().Mutex().RUnlock()
	if ok {
		oldParamMap := oldParam.AsValueMap()
		for k, v := range oldParamMap {
			params[k] = v
		}
	}

	// only if the stage has a use block, we will use the parameters from the use block
	if s.Use != nil && s.Use.Parameters != nil {
		s.Eval().Mutex().RLock()
		parameters, d := s.Use.Parameters.Value(s.Eval().Context())
		s.Eval().Mutex().RUnlock()
		diags = diags.Extend(d)
		if !parameters.IsNull() {
			for k, v := range parameters.AsValueMap() {
				params[k] = v
			}
		}
	}

	s.eval.evalContext = s.eval.evalContext.NewChild()

	// we do not need to use a lock here because we are modifying the map
	// on the child evalContext, not the host evalContext
	s.Eval().Context().Variables = map[string]cty.Value{}
	s.Eval().Context().Variables[blocks.ParamBlock] = cty.ObjectVal(params)
	return s
}

// expandSelf expands the properties of the stage into the context
// such as the name, id, hook, and status
func (s *stageConductor) expandSelf() *stageConductor {
	id := s.Id
	name := s.Name
	if s.cfg.Parent != nil {
		s.Logger().Debugf("using parent %s.%s", s.cfg.Parent.Name, s.cfg.Parent.Id)
		id = s.cfg.Parent.Id
		name = s.cfg.Parent.Name
	}
	s.eval.evalContext = s.eval.evalContext.NewChild()
	s.Eval().Context().Variables = map[string]cty.Value{
		ThisBlock: cty.ObjectVal(map[string]cty.Value{
			"name":   cty.StringVal(name),
			"id":     cty.StringVal(id),
			"hook":   cty.BoolVal(s.cfg.Hook),
			"status": cty.StringVal(string(s.cfg.Status.Status)),
		}),
	}
	return s
}

// expandEach expands each block into the context
// includes values such as each.key and each.value if the block has a ForEach attribute
func (s *stageConductor) expandEach() *stageConductor {
	s.eval.evalContext = s.eval.evalContext.NewChild()
	if s.cfg.Each != nil {
		s.Eval().Context().Variables = map[string]cty.Value{
			EachBlock: cty.ObjectVal(s.cfg.Each),
		}
	}
	return s
}
