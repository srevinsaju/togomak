package ci

import (
	"github.com/hashicorp/hcl/v2"
	dataBlock "github.com/srevinsaju/togomak/v1/pkg/blocks/data"
)

func (d Data) Variables() []hcl.Traversal {
	var traversal []hcl.Traversal
	provider := dataBlock.DefaultProviders.Get(d.Provider)
	provide := provider.New()
	traversal = append(traversal, dataBlock.Variables(provide, d.Body)...)
	return traversal
}

func (d Datas) Variables() []hcl.Traversal {
	var traversal []hcl.Traversal
	for _, data := range d {
		traversal = append(traversal, data.Variables()...)
	}
	return traversal
}
