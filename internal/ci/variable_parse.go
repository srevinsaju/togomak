package ci

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
	"strings"
)

func ParseVariableShell(raw string) (*Variable, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	eq := strings.Index(raw, "=")
	if eq == -1 {
		diags = diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid -var option",
			Detail:   fmt.Sprintf("The given -var option %q is not correctly specified. Must be a variable name and value separated by an equals sign, like -var=\"key=value\".", raw),
		})
		return nil, diags
	}
	name := raw[:eq]
	rawVal := raw[eq+1:]
	if strings.HasSuffix(name, " ") {
		diags = diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid -var option",
			Detail:   fmt.Sprintf("Variable name %q is invalid due to trailing space. Did you mean -var=\"%s=%s\"?", name, strings.TrimSuffix(name, " "), strings.TrimPrefix(rawVal, " ")),
		})
		return nil, diags
	}
	return &Variable{
		Id:    name,
		Value: hcl.StaticExpr(cty.StringVal(rawVal), hcl.Range{}),
	}, diags

}
