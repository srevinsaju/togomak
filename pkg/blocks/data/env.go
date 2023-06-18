package data

import (
	"context"
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/srevinsaju/togomak/v1/pkg/c"
	"github.com/srevinsaju/togomak/v1/pkg/diag"
	"github.com/zclconf/go-cty/cty"
	"os"
)

type EnvProvider struct {
	initialized bool
	Key         hcl.Expression `hcl:"key" json:"key"`
	Default     hcl.Expression `hcl:"default" json:"default"`

	keyParsed string
	def       string
	ctx       context.Context
}

func (e *EnvProvider) Name() string {
	return "env"
}

func (e *EnvProvider) Initialized() bool {
	return e.initialized
}

func (e *EnvProvider) SetContext(context context.Context) {
	if !e.initialized {
		panic("provider not initialized")
	}
	e.ctx = context
}

func (e *EnvProvider) Version() string {
	return "1"
}

func (e *EnvProvider) New() Provider {
	return &EnvProvider{
		initialized: true,
	}
}

func (e *EnvProvider) Attributes(ctx context.Context) map[string]cty.Value {
	return map[string]cty.Value{
		"key":     cty.StringVal(e.keyParsed),
		"default": cty.StringVal(e.def),
	}
}

func (e *EnvProvider) Url() string {
	return "embedded::togomak.srev.in/providers/data/env"
}

func (e *EnvProvider) Schema() *hcl.BodySchema {
	schema := &hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{
			{
				Name:     "key",
				Required: true,
			},
			{
				Name:     "default",
				Required: false,
			},
		},
	}
	return schema

}

func (e *EnvProvider) DecodeBody(body hcl.Body) diag.Diagnostics {
	if !e.initialized {
		panic("provider not initialized")
	}
	var diags diag.Diagnostics
	hclDiagWriter := e.ctx.Value(c.TogomakContextHclDiagWriter).(hcl.DiagnosticWriter)
	hclContext := e.ctx.Value(c.TogomakContextHclEval).(*hcl.EvalContext)

	schema := e.Schema()
	content, hclDiags := body.Content(schema)
	if hclDiags.HasErrors() {
		source := fmt.Sprintf("data.%s:decodeBody", e.Name())
		diags = diags.NewHclWriteDiagnosticsError(source, hclDiagWriter.WriteDiagnostics(hclDiags))
	}
	attr := content.Attributes["key"]
	var key cty.Value
	key, hclDiags = attr.Expr.Value(hclContext)
	if hclDiags.HasErrors() {
		source := fmt.Sprintf("data.%s:decodeBody", e.Name())
		diags = diags.NewHclWriteDiagnosticsError(source, hclDiagWriter.WriteDiagnostics(hclDiags))
	}
	e.keyParsed = key.AsString()

	attr = content.Attributes["default"]
	key, hclDiags = attr.Expr.Value(hclContext)
	if hclDiags.HasErrors() {
		source := fmt.Sprintf("data.%s:decodeBody", e.Name())
		diags = diags.NewHclWriteDiagnosticsError(source, hclDiagWriter.WriteDiagnostics(hclDiags))
	}
	e.def = key.AsString()

	return diags

}

func (e *EnvProvider) Value(ctx context.Context, id string) string {
	if !e.initialized {
		panic("provider not initialized")
	}
	v, exists := os.LookupEnv(e.keyParsed)
	if exists {
		return v
	}
	return e.def
}
