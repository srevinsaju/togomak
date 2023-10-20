package ci

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

type Locals struct {
	Body hcl.Body `hcl:",remain"`
}

type Local struct {
	Key   string
	Value hcl.Expression `hcl:"value,optional"`
	value cty.Value
}

type LocalsGroup []Locals
type LocalGroup []*Local
