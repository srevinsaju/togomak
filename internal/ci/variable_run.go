package ci

import (
	"context"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/ext/typeexpr"
	"github.com/srevinsaju/togomak/v1/internal/blocks"
	"github.com/srevinsaju/togomak/v1/internal/global"
	"github.com/srevinsaju/togomak/v1/internal/runnable"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
	"os"
)

func (v *Variable) Prepare(conductor *Conductor, skip bool, overridden bool) hcl.Diagnostics {
	return nil // no-op
}

func (v *Variable) resolveVar(conductor *Conductor) (cty.Value, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	osEnvValue := os.Getenv(fmt.Sprintf("TOGOMAK_VAR_%s", v.Id))
	if osEnvValue != "" {
		return cty.StringVal(osEnvValue), nil
	}
	for _, cliVariable := range conductor.Variables() {
		if cliVariable.Id == v.Id {
			conductor.Eval().Mutex().RLock()
			b, d := cliVariable.Value.Value(conductor.Eval().Context())
			conductor.Eval().Mutex().RUnlock()
			return b, diags.Extend(d)
		}
	}
	if v.Default != nil {
		conductor.Eval().Mutex().RLock()
		def, d := v.Default.Value(conductor.Eval().Context())
		conductor.Eval().Mutex().RUnlock()
		if d.HasErrors() {
			return cty.NilVal, d
		}
		if !def.IsNull() && def.IsKnown() {
			return def, nil
		}
	}
	var resp string
	conductor.StdinLock()
	defer conductor.StdinUnlock()
	err := survey.AskOne(&survey.Input{
		Message: fmt.Sprintf("%s.%s", blocks.VarBlock, v.Id),
		Default: "",
		Help:    v.Desc,
	}, &resp)
	if err != nil || resp == "" {
		return cty.NilVal, diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "No value for required variable",
			Detail:   fmt.Sprintf("The root module input variable \"%s\" is not set, and has no default value. Use a -var, -var-file or a TOGOMAK_VAR_%s command line argument to provide a value for this variable.", v.Id, v.Id),
		})
	} else {
		return cty.StringVal(resp), diags
	}
}

func (v *Variable) resolveVarTypedWithDefaults(conductor *Conductor) (cty.Value, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	value, d := v.resolveVar(conductor)
	diags = diags.Extend(d)

	// the user specified type, we will check
	// if this type is null, or undefined, and include
	// a fallback type
	conductor.Eval().Mutex().RLock()
	uType, d := v.Ty.Value(conductor.Eval().Context())
	conductor.Eval().Mutex().RUnlock()

	ty := cty.DynamicPseudoType
	var def *typeexpr.Defaults

	if !d.HasErrors() && !uType.IsNull() {
		ty, def, d = typeexpr.TypeConstraintWithDefaults(v.Ty)
		diags = diags.Extend(d)
	}

	// we will first apply the default type constraints
	// only if the default is not nil and the value inferred from the command line or the
	// environment variable is not null
	// https://github.com/hashicorp/terraform/blob/e539f46e852f2dad6c35de4c8d09dc6b04c9eafb/internal/configs/named_values.go#L153
	if def != nil && !value.IsNull() {
		value = def.Apply(value)
	}

	value, err := convert.Convert(value, ty)
	if err != nil {
		diags = diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid variable value",
			Detail:   fmt.Sprintf("The value of variable %s is invalid: %s", v.Id, err),
		})
		return cty.NilVal, diags
	}
	return value, diags
}

func (v *Variable) Run(conductor *Conductor, options ...runnable.Option) (diags hcl.Diagnostics) {
	// logger := conductor.Logger().WithField("var", v.Id)
	// cfg := runnable.NewConfig(options...)
	evalContext := conductor.Eval().Context()

	value, d := v.resolveVarTypedWithDefaults(conductor)
	diags = diags.Extend(d)
	if diags.HasErrors() {
		return diags
	}

	global.VariableBlockEvalContextMutex.Lock()
	conductor.Eval().Mutex().RLock()
	data, ok := evalContext.Variables[blocks.VarBlock]
	conductor.Eval().Mutex().RUnlock()
	dataMutated := map[string]cty.Value{}
	if ok {
		for k, val := range data.AsValueMap() {
			dataMutated[k] = val
		}
	}
	dataMutated[v.Id] = value
	conductor.Eval().Mutex().Lock()
	evalContext.Variables[blocks.VarBlock] = cty.ObjectVal(dataMutated)
	conductor.Eval().Mutex().Unlock()
	global.VariableBlockEvalContextMutex.Unlock()
	// endregion

	if diags.HasErrors() {
		return diags
	}
	return nil
}

func (v *Variable) CanRun(conductor *Conductor, options ...runnable.Option) (bool, hcl.Diagnostics) {
	return true, nil
}

func (v *Variable) Terminated() bool {
	return true
}

func (v *Variable) Terminate(conductor *Conductor, safe bool) hcl.Diagnostics {
	return nil
}

func (v *Variable) Kill() hcl.Diagnostics {
	return nil
}

func (v *Variable) ExecutionOptions(ctx context.Context) (*DaemonLifecycleConfig, hcl.Diagnostics) {
	return nil, nil
}
