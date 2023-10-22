package data

import (
	"context"
	"github.com/hashicorp/hcl/v2"
	"github.com/sirupsen/logrus"
	"github.com/srevinsaju/togomak/v1/internal/conductor"
	"github.com/zclconf/go-cty/cty"
	"sync"
)

type Provider interface {
	Name() string
	Url() string
	Version() string
	Schema() *hcl.BodySchema
	Initialized() bool
	New() Provider

	SetContext(context context.Context)
	DecodeBody(conductor conductor.Conductor, body hcl.Body, opts ...ProviderOption) hcl.Diagnostics
	Value(conductor conductor.Conductor, ctx context.Context, id string, opts ...ProviderOption) (string, hcl.Diagnostics)
	Attributes(conductor conductor.Conductor, ctx context.Context, id string, opts ...ProviderOption) (map[string]cty.Value, hcl.Diagnostics)
}

type Eval struct {
	context *hcl.EvalContext
	mu      *sync.RWMutex
}

func (e *Eval) Context() *hcl.EvalContext {
	return e.context
}

func (e *Eval) Mutex() *sync.RWMutex {
	return e.mu
}

type Conductor struct {
	eval   *Eval
	logger *logrus.Entry
}

func (c *Conductor) Eval() *hcl.EvalContext {
	return c.eval.context
}

func (c *Conductor) Logger() *logrus.Entry {
	return c.logger
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
