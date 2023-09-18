package data

import (
	"context"
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/srevinsaju/togomak/v1/pkg/global"
	"github.com/zclconf/go-cty/cty"
	"os"
)

const (
	EnvProviderAttrKey     = "key"
	EnvProviderAttrDefault = "default"
)

type EnvProvider struct {
	initialized bool
	Key         hcl.Expression `hcl:"key" json:"key"`
	Default     hcl.Expression `hcl:"default" json:"default"`

	keyParsed string
	def       string
	defOk     bool
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

func (e *EnvProvider) Attributes(ctx context.Context, id string) (map[string]cty.Value, hcl.Diagnostics) {
	return map[string]cty.Value{
		EnvProviderAttrKey:     cty.StringVal(e.keyParsed),
		EnvProviderAttrDefault: cty.StringVal(e.def),
	}, nil
}

func (e *EnvProvider) Url() string {
	return "embedded::togomak.srev.in/providers/data/env"
}

func (e *EnvProvider) Schema() *hcl.BodySchema {
	schema := &hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{
			{
				Name:     EnvProviderAttrKey,
				Required: true,
			},
			{
				Name:     EnvProviderAttrDefault,
				Required: false,
			},
		},
	}
	return schema

}

func (e *EnvProvider) DecodeBody(body hcl.Body) hcl.Diagnostics {
	if !e.initialized {
		panic("provider not initialized")
	}
	var diags hcl.Diagnostics
	hclContext := global.HclEvalContext()

	schema := e.Schema()
	content, d := body.Content(schema)
	diags = diags.Extend(d)

	attr := content.Attributes["key"]
	var key cty.Value

	global.EvalContextMutex.RLock()
	key, d = attr.Expr.Value(hclContext)
	global.EvalContextMutex.RUnlock()
	diags = diags.Extend(d)

	e.keyParsed = key.AsString()

	attr, ok := content.Attributes["default"]
	if !ok {
		e.defOk = false
		e.def = ""
		return diags
	}

	global.EvalContextMutex.RLock()
	key, d = attr.Expr.Value(hclContext)
	global.EvalContextMutex.RUnlock()
	diags = diags.Extend(d)

	e.def = key.AsString()
	e.defOk = true

	return diags

}

func (e *EnvProvider) Value(ctx context.Context, id string) (string, hcl.Diagnostics) {
	if !e.initialized {
		panic("provider not initialized")
	}
	v, exists := os.LookupEnv(e.keyParsed)
	if exists {
		return v, nil
	}
	if !e.defOk {
		return "", hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "environment variable not found",
				Detail:   fmt.Sprintf("environment variable %s not found", e.keyParsed),
			},
		}
	}
	return e.def, nil
}
