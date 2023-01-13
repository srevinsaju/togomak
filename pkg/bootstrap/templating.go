package bootstrap

import (
	"github.com/spf13/afero"
	"github.com/srevinsaju/togomak/pkg/context"
	"strings"
)

func Templating(ctx *context.Context) {
	ctx.Data["path"] = map[string]interface{}{

		"exists": func(path string) (bool, error) {
			return afero.Exists(afero.NewOsFs(), path)
		},
	}
	ctx.Data["strings"] = map[string]interface{}{
		"contains": func(str string, substr string) bool {
			return strings.Contains(str, substr)
		},
		"hasPrefix": func(str string, substr string) bool {
			return strings.HasPrefix(str, substr)
		},
		"hasSuffix": func(str string, substr string) bool {
			return strings.HasSuffix(str, substr)
		},
		"split": func(str string, substr string) []string {
			return strings.Split(str, substr)
		},
		"join": func(str []string, substr string) string {
			return strings.Join(str, substr)
		},
		"replace": func(str string, substr string, repl string) string {
			return strings.Replace(str, substr, repl, -1)
		},
		"trim": func(str string, substr string) string {
			return strings.Trim(str, substr)
		},
		"trimLeft": func(str string, substr string) string {
			return strings.TrimLeft(str, substr)
		},
		"trimRight": func(str string, substr string) string {
			return strings.TrimRight(str, substr)
		},
	}
}
