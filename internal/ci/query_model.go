package ci

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/sirupsen/logrus"
	"github.com/srevinsaju/togomak/v1/internal/blocks"
	"github.com/srevinsaju/togomak/v1/internal/global"
	"github.com/zclconf/go-cty/cty"
	"strings"
)

type QueryEngine struct {
	Rule   hcl.Expression
	rule   string
	Logger *logrus.Entry

	empty bool
}

func New(comp string) (*QueryEngine, hcl.Diagnostics) {
	var diags hcl.Diagnostics

	logger := global.Logger().WithField("rules", "")
	logger.Debugf("received rule '%s'", comp)
	if comp == "" || strings.Trim(comp, " ") == "" {
		return &QueryEngine{
			empty:  true,
			rule:   comp,
			Logger: logger,
		}, diags
	}

	p, d := hclsyntax.ParseExpression([]byte(comp), "<input>", hcl.Pos{Line: 1, Column: 1})
	diags = diags.Extend(d)

	return &QueryEngine{
		Rule:   p,
		rule:   comp,
		Logger: logger,
	}, diags
}

func (e *QueryEngine) Eval(conductor *Conductor, ok bool, stage PhasedBlock) (bool, bool, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	var d hcl.Diagnostics
	if e.empty {
		return ok, false, diags
	}

	ectx := conductor.Eval().Context()
	ectx = ectx.NewChild()
	ectx.Variables = map[string]cty.Value{}
	ectx.Variables["if"] = cty.BoolVal(ok)
	ectx.Variables["id"] = cty.StringVal(stage.Identifier())
	ectx.Variables["name"] = cty.StringVal(stage.Description().Name)

	lifecyclePhase := cty.ListVal([]cty.Value{cty.StringVal("default")})
	lifecycleTimeout := cty.NumberIntVal(0)
	if stage.LifecycleConfig() != nil {
		conductor.Eval().Mutex().RLock()
		lifecyclePhase, d = stage.LifecycleConfig().Phase.Value(ectx)
		conductor.Eval().Mutex().RUnlock()

		diags = diags.Extend(d)

		conductor.Eval().Mutex().RLock()
		lifecycleTimeout, d = stage.LifecycleConfig().Timeout.Value(ectx)
		conductor.Eval().Mutex().RUnlock()
		diags = diags.Extend(d)
	}

	ectx.Variables[blocks.LifecycleBlock] = cty.ObjectVal(map[string]cty.Value{
		"phase":   lifecyclePhase,
		"timeout": lifecycleTimeout,
	})

	conductor.Eval().Mutex().RLock()
	v, d := e.Rule.Value(ectx)
	conductor.Eval().Mutex().RUnlock()
	diags = diags.Extend(d)

	if v.Type() != cty.Bool {
		diags = diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Rule must be a boolean expression",
			Detail:   "The rule must be a boolean expression, but the given expression has type " + v.Type().FriendlyName(),
		})
		return false, true, diags
	}
	if !v.IsKnown() || !v.IsWhollyKnown() {
		diags = diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Rule must be a known value",
			Detail:   "The rule must be a known value, but the given expression has unknown parts",
		})
	}

	e.Logger.WithField("runnable", stage.Identifier()).Debugf("evaluated rule '%s' to %v", e.rule, v.True())
	return v.True(), true, diags
}
