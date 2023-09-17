package ci

import "github.com/hashicorp/hcl/v2"

type Import struct {
	id     string
	Source hcl.Expression `hcl:"source" json:"source"`
}

type Imports []Import
