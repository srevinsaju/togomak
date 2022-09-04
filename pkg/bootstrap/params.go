package bootstrap

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/flosch/pongo2/v6"
	"github.com/mattn/go-isatty"
	"github.com/srevinsaju/togomak/pkg/context"
	"github.com/srevinsaju/togomak/pkg/schema"
	"os"
)

func Params(ctx *context.Context, data schema.SchemaConfig) {
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
			if isatty.IsTerminal(os.Stdin.Fd()) {
				prompt := &survey.Input{
					Message: fmt.Sprintf("Enter the value for param.%s: ", param.Name),
				}
				err := survey.AskOne(prompt, &v)
				if err != nil {
					panic(err)
				}
			} else {
				ctx.Logger.Fatalf("The parameter '%s' does not have a value, use TOGOMAK__%s to set it, or use a default value in parameters[name=%s].default", param.Name, param.Name, param.Name)
			}

		}
		paramCtx.Data[param.Name] = v
	}

}
