package ci

import "github.com/hashicorp/hcl/v2"

type Module struct {
	Id        string         `hcl:"id,label" json:"id"`
	DependsOn hcl.Expression `hcl:"depends_on,optional" json:"depends_on"`

	Source hcl.Expression `hcl:"source" json:"source"`

	pipeline *Pipeline
}

type Modules []Module
