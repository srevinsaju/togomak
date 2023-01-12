package templating

import (
	"github.com/flosch/pongo2/v6"
	"github.com/spf13/afero"
	"github.com/srevinsaju/togomak/pkg/schema"
	"os"
)

func Env(name *pongo2.Value) *pongo2.Value {
	return pongo2.AsValue(os.Getenv(name.String()))
}

func Execute(tpl *pongo2.Template, data map[string]interface{}) (string, error) {

	data["path"] = map[string]interface{}{

		"exists": func(path string) (bool, error) {
			return afero.Exists(afero.NewOsFs(), path)
		},
	}
	return tpl.Execute(data)
}

func ExecuteWithStage(tpl *pongo2.Template, data map[string]interface{}, stage schema.StageConfig) (string, error) {
	data["id"] = stage.Id
	data["name"] = stage.Name
	data["description"] = stage.Description

	return Execute(tpl, data)
}
