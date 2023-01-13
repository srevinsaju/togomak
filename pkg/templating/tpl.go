package templating

import (
	"github.com/flosch/pongo2/v6"
	"github.com/srevinsaju/togomak/pkg/schema"
	"os"
)

func Env(name *pongo2.Value) *pongo2.Value {
	return pongo2.AsValue(os.Getenv(name.String()))
}

func cloneMap(data map[string]interface{}) map[string]interface{} {
	// clone the data map to prevent data pollution
	// from the template
	newData := make(map[string]interface{})

	for k, v := range data {
		newData[k] = v
	}
	return newData
}

func Execute(tpl *pongo2.Template, data map[string]interface{}) (string, error) {
	// clone the data map to prevent data pollution
	// from the template
	return tpl.Execute(data)
}

func ExecuteWithStage(tpl *pongo2.Template, data map[string]interface{}, stage schema.StageConfig) (string, error) {
	data = cloneMap(data)

	data["id"] = stage.Id
	data["name"] = stage.Name
	data["description"] = stage.Description

	return Execute(tpl, data)
}
