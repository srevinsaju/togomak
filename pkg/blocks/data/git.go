package data

import (
	"context"
	"github.com/hashicorp/hcl/v2"
	"github.com/srevinsaju/togomak/v1/pkg/diag"
)

type GitProvider struct {
	initialized bool
	Default     hcl.Expression `hcl:"default" json:"default"`

	ctx context.Context
}

func (e *GitProvider) Name() string {
	return "git"
}

func (e *GitProvider) SetContext(context context.Context) {
	e.ctx = context
}

func (e *GitProvider) Version() string {
	return "1"
}

func (e *GitProvider) Url() string {
	return "embedded::togomak.srev.in/providers/data/git"
}

func (e *GitProvider) DecodeBody(body hcl.Body) diag.Diagnostics {
	if !e.initialized {
		panic("provider not initialized")
	}
	/*var diags diag.Diagnostics
	hclDiagWriter := e.ctx.Value(c.TogomakContextHclDiagWriter).(hcl.DiagnosticWriter)
	// hclContext := e.ctx.Value(c.TogomakContextHclEval).(*hcl.EvalContext)

	schema := e.Schema()
	content, hclDiags := body.Content(schema)
	if hclDiags.HasErrors() {
		source := fmt.Sprintf("data.%s:decodeBody", e.Name())
		diags = diags.NewHclWriteDiagnosticsError(source, hclDiagWriter.WriteDiagnostics(hclDiags))
	}
	url := content.Attributes["url"]*/
	return nil
}

func (e *GitProvider) New() Provider {
	return &GitProvider{
		initialized: true,
	}
}

func (e *GitProvider) Schema() *hcl.BodySchema {
	return &hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{
			{
				Name:     "url",
				Required: true,
			},
			{
				Name:     "tag",
				Required: false,
			},
			{
				Name:     "branch",
				Required: false,
			},
			{
				Name:     "commit",
				Required: false,
			},
		},
	}
}

func (e *GitProvider) Initialized() bool {
	return e.initialized
}

func (e *GitProvider) Value() string {
	if !e.initialized {
		panic("provider not initialized")
	}

	return ""
}
