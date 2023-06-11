package ci

import "github.com/hashicorp/hcl/v2"

const DataProviderBlock = "provider"

type DataProvider struct {
	Name string `hcl:"name,label" json:"name"`
	Url  string `hcl:"url,optional" json:"url,omitempty"`
}

func (d DataProvider) Variables() []hcl.Traversal {
	return nil
}

type DataProviders []DataProvider

func (d DataProviders) Variables() []hcl.Traversal {
	var traversal []hcl.Traversal
	for _, provider := range d {
		traversal = append(traversal, provider.Variables()...)
	}
	return traversal
}
