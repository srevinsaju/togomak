package data

import (
	"context"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/hashicorp/hcl/v2"
	"github.com/sirupsen/logrus"
	"github.com/srevinsaju/togomak/v1/pkg/c"
	"github.com/srevinsaju/togomak/v1/pkg/diag"
	"github.com/srevinsaju/togomak/v1/pkg/meta"
	"github.com/zclconf/go-cty/cty"
	"os"
)

type PromptProvider struct {
	initialized bool

	Prompt  hcl.Expression `hcl:"prompt" json:"prompt"`
	Default hcl.Expression `hcl:"default" json:"default"`

	promptParsed string
	def          string
	ctx          context.Context
}

func (e *PromptProvider) Name() string {
	return "prompt"
}

func (e *PromptProvider) SetContext(context context.Context) {
	e.ctx = context
}

func (e *PromptProvider) Version() string {
	return "1"
}

func (e *PromptProvider) Url() string {
	return "embedded::togomak.srev.in/providers/data/prompt"
}

func (e *PromptProvider) DecodeBody(body hcl.Body) diag.Diagnostics {
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
	attr := content.Attributes["prompt"]
	var key cty.Value
	key, hclDiags = attr.Expr.Value(hclContext)
	if hclDiags.HasErrors() {
		source := fmt.Sprintf("data.%s:decodeBody", e.Name())
		diags = diags.NewHclWriteDiagnosticsError(source, hclDiagWriter.WriteDiagnostics(hclDiags))
	}
	e.promptParsed = key.AsString()

	attr = content.Attributes["default"]
	key, hclDiags = attr.Expr.Value(hclContext)
	if hclDiags.HasErrors() {
		source := fmt.Sprintf("data.%s:decodeBody", e.Name())
		diags = diags.NewHclWriteDiagnosticsError(source, hclDiagWriter.WriteDiagnostics(hclDiags))
	}
	e.def = key.AsString()

	return diags

}

func (e *PromptProvider) New() Provider {
	return &PromptProvider{
		initialized: true,
	}
}

func (e *PromptProvider) Schema() *hcl.BodySchema {
	return &hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{
			{
				Name:     "prompt",
				Required: false,
			},
			{
				Name:     "default",
				Required: false,
			},
		},
	}
}

func (e *PromptProvider) Attributes(ctx context.Context) map[string]cty.Value {
	return map[string]cty.Value{
		"prompt":  cty.StringVal(e.promptParsed),
		"default": cty.StringVal(e.def),
	}
}

func (e *PromptProvider) Initialized() bool {
	return e.initialized
}

func (e *PromptProvider) Value(ctx context.Context, id string) string {
	if !e.initialized {
		panic("provider not initialized")
	}

	logger := e.ctx.Value(c.TogomakContextLogger).(*logrus.Logger).WithField("provider", e.Name())
	unattended := e.ctx.Value(c.TogomakContextUnattended).(bool)

	envVarName := fmt.Sprintf("%s%s__%s", meta.EnvVarPrefix, e.Name(), id)
	logger.Tracef("checking for environment variable %s", envVarName)
	envExists, ok := os.LookupEnv(envVarName)
	if ok {
		logger.Debug("environment variable found, using that")
		return envExists
	}
	if unattended {
		logger.Warn("--unattended/--ci mode enabled, falling back to default")
		return e.def
	}

	prompt := e.promptParsed
	if prompt == "" {
		prompt = fmt.Sprintf("Enter a value for %s:", e.Name())
	}

	input := survey.Input{
		Renderer: survey.Renderer{},
		Message:  prompt,
		Default:  e.def,
		Help:     "",
		Suggest:  nil,
	}
	var resp string
	err := survey.AskOne(&input, &resp)
	if err != nil {
		logger.Warn("unable to get value from prompt: ", err)
		return e.def
	}

	return resp
}
