package bootstrap

import (
	"bytes"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/flosch/pongo2/v6"
	"github.com/hashicorp/go-envparse"
	"github.com/mattn/go-isatty"
	"github.com/srevinsaju/togomak/pkg/config"
	"github.com/srevinsaju/togomak/pkg/context"
	"github.com/srevinsaju/togomak/pkg/schema"
	"os"
	"strings"
)

func Params(ctx *context.Context, data schema.SchemaConfig, noInteractive bool) {
	paramCtx := ctx.AddChild("param", "")

	for _, param := range data.Parameters {
		v := ""
		if paramCtx.Getenv(param.Name) != "" {
			v = paramCtx.Getenv(param.Name)
		}
		if param.Default != "" {
			tpl, err := pongo2.FromString(param.Default)
			if err != nil {
				paramCtx.Logger.Fatal("Cannot render args:", err)
			}
			parsed, err := tpl.Execute(ctx.Data.AsMap())
			if err != nil {
				paramCtx.Logger.Fatal("Cannot render args:", err)
			}
			v = parsed
		}
		if v == "" {
			// prompt the user for the value
			if isatty.IsTerminal(os.Stdin.Fd()) && !noInteractive {
				prompt := &survey.Input{
					Message: fmt.Sprintf("Enter the value for param.%s: ", param.Name),
				}
				err := survey.AskOne(prompt, &v)
				if err != nil {
					panic(err)
				}
			}

		}
		paramCtx.DataMutex.Lock()
		paramCtx.Data[param.Name] = v
		paramCtx.DataMutex.Unlock()
	}

}

func VerifyParams(ctx *context.Context, data schema.SchemaConfig) {
	paramCtx := ctx.AddChild("param", "")

	for _, param := range data.Parameters {
		if paramCtx.Getenv(param.Name) == "" {
			paramCtx.Logger.Warnf("The parameter '%s' does not have a value, use TOGOMAK__%s to set it, or use a default value in parameters[name=%s].default", param.Name, param.Name, param.Name)
			paramCtx.Logger.Fatalf("param.%s is not set", param.Name)
		}
	}
}

func OverrideParams(ctx *context.Context, cfg config.Config) {
	paramCtx := ctx.AddChild("param", "")

	builder := strings.Builder{}

	for _, param := range cfg.Parameters {
		// param is in the format k=v
		// split it
		paramCtx.Logger.Tracef("received raw parameter %s", param)
		builder.WriteString(param)
		builder.WriteString("\n")
	}

	// parse the parameters
	// and override the parameters
	// in the context
	m, err := envparse.Parse(bytes.NewReader([]byte(builder.String())))
	if err != nil {
		paramCtx.Logger.Fatal("cannot parse parameters:", err)
		panic(err)
	}
	for k, v := range m {
		paramCtx.Data[k] = v
	}
}
