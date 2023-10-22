package ci

import (
	"github.com/hashicorp/hcl/v2"
	dataBlock "github.com/srevinsaju/togomak/v1/internal/blocks/data"
)

func (s *Data) Variables() []hcl.Traversal {
	var traversal []hcl.Traversal
	provider := dataBlock.DefaultProviders.Get(s.Provider)
	// TODO: this will panic, if the provider is not found
	if provider == nil {
		panic("provider not found")
	}
	provide := provider.New()
	traversal = append(traversal, dataBlock.Variables(provide, s.Body)...)
	return traversal
}

func (d Datas) Variables() []hcl.Traversal {
	var traversal []hcl.Traversal
	for _, data := range d {
		traversal = append(traversal, data.Variables()...)
	}
	return traversal
}
