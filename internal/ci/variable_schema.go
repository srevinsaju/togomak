package ci

import "github.com/hashicorp/hcl/v2"

type Variable struct {
	Id string `hcl:"id,label" json:"id"`

	Desc      string         `hcl:"description,optional" json:"description"`
	DependsOn hcl.Expression `hcl:"depends_on,optional" json:"depends_on"`
	Value     hcl.Expression `hcl:"value,optional" json:"value"`
	Default   hcl.Expression `hcl:"default,optional" json:"default"`
	Ty        hcl.Expression `hcl:"type,optional" json:"type"`
}

type Variables []*Variable
