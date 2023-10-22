package ci

import "github.com/hashicorp/hcl/v2"

type Data struct {
	Provider string `hcl:"provider,label" json:"provider"`
	Id       string `hcl:"id,label" json:"id"`

	Name  string `hcl:"name,optional" json:"name"`
	Value string `json:"value"`

	Body hcl.Body `hcl:",remain"`
}

type Datas []Data
