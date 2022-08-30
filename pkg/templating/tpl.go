package templating

import (
	"github.com/flosch/pongo2/v6"
	"os"
)

func Env(name *pongo2.Value) *pongo2.Value {
	return pongo2.AsValue(os.Getenv(name.String()))
}
