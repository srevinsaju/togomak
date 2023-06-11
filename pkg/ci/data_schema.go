package ci

import "github.com/hashicorp/hcl/v2"

type Data struct {
	Provider string `hcl:"provider,label" json:"provider"`
	Id       string `hcl:"id,label" json:"id"`

	Value string `json:"value"`

	Body hcl.Body `hcl:",remain"`
}

type Datas []Data
