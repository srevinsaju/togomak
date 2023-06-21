package data

import (
	"context"
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

type Provider interface {
	Name() string
	Url() string
	Version() string
	Schema() *hcl.BodySchema
	Initialized() bool
	New() Provider

	SetContext(context context.Context)
	DecodeBody(body hcl.Body) hcl.Diagnostics
	Value(ctx context.Context, id string) (string, hcl.Diagnostics)
	Attributes(ctx context.Context) (map[string]cty.Value, hcl.Diagnostics)
}

func Variables(e Provider, body hcl.Body) []hcl.Traversal {

	if !e.Initialized() {
		panic("provider not initialized")
	}
	var traversal []hcl.Traversal

	schema := e.Schema()
	d, _, diags := body.PartialContent(schema)
	if diags.HasErrors() {
		panic(diags.Error())
	}
	for _, attr := range d.Attributes {
		traversal = append(traversal, attr.Expr.Variables()...)
	}
	return traversal
}
