package ci

import "github.com/hashicorp/hcl/v2"

type Locals struct {
	Body hcl.Body `hcl:",remain"`
}

type LocalsGroup []Locals
