package ci

import (
	"context"
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/srevinsaju/togomak/v1/pkg/c"
	"github.com/srevinsaju/togomak/v1/pkg/diag"
	"github.com/zclconf/go-cty/cty"
)

func (s *Stage) Description() string {
	// TODO: implement
	return ""
}

func (s *Stage) Identifier() string {
	if s.forEachAttr != "" {
		return fmt.Sprintf("%s.%s[\"%s\"]", StageBlock, s.Id, s.forEachAttr)
	}
	return fmt.Sprintf("%s.%s", StageBlock, s.Id)
}

func (s *Stage) Type() string {
	return StageBlock
}

func (s *Stage) IsDaemon() bool {
	return s.Daemon != nil && s.Daemon.Enabled
}

func (s Stages) ById(id string) (*Stage, error) {
	for _, stage := range s {
		if stage.Id == id {
			return &stage, nil
		}
	}
	return nil, fmt.Errorf("stage with id %s not found", id)
}

func (s *Stage) Expand(ctx context.Context) (Runnables, diag.Diagnostics) {
	var diags diag.Diagnostics
	evalCtx := ctx.Value(c.TogomakContextHclEval).(*hcl.EvalContext)
	hclDiagnosticWriter := ctx.Value(c.TogomakContextHclDiagWriter).(hcl.DiagnosticWriter)

	// expand stages using macros
	s, d := s.expandMacros(ctx)
	if d.HasErrors() {
		return nil, diags.ExtendHCLDiagnostics(d, hclDiagnosticWriter, s.Identifier())
	}

	v, d := s.ForEach.Value(evalCtx)
	if d.HasErrors() {
		return nil, diags.ExtendHCLDiagnostics(d, hclDiagnosticWriter, s.Identifier())
	}
	if v.IsNull() {
		return nil, diags
	}

	fmt.Println(v.LengthInt())
	// https://github.com/hashicorp/hcl/blob/7208bce57fadb72db3a328ebc9aa86489cd06fce/ext/dynblock/expand_spec.go#LL47C1-L48C1
	if !v.CanIterateElements() && v.Type() != cty.DynamicPseudoType {
		return nil, diags.ExtendHCLDiagnostics(hcl.Diagnostics{
			{
				Severity:    hcl.DiagError,
				Summary:     "Invalid type",
				Detail:      "The value of the `for_each` argument must be a map, or a set of strings.",
				Subject:     s.ForEach.Range().Ptr(),
				Context:     s.ForEach.Range().Ptr(),
				EvalContext: evalCtx,
			},
		}, hclDiagnosticWriter, s.Identifier())
	}
	slice := v.AsValueMap()

	var stages Runnables
	for i, v := range slice {
		v := v
		stages = append(stages, &Stage{
			Id:             s.Id,
			Condition:      s.Condition,
			DependsOn:      s.DependsOn,
			Use:            s.Use,
			ForEach:        nil,
			forEachAttr:    i,
			forEachValue:   v,
			isForEachBlock: false,
			Daemon:         s.Daemon,
			Retry:          s.Retry,
			Name:           s.Name,
			Dir:            s.Dir,
			Script:         s.Script,
			Args:           s.Args,
			Container:      s.Container,
			Environment:    s.Environment,
			process:        s.process,
			ContainerId:    s.ContainerId,
		})
	}
	s.isForEachBlock = true
	return stages, diags
}

func (s *Stage) Expanded() bool {
	return s.isForEachBlock
}

func (s *Stage) ForEachDerived() bool {
	return s.forEachAttr != ""
}
