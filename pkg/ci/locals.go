package ci

import (
	"fmt"
	"github.com/srevinsaju/togomak/v1/pkg/diag"
)

const LocalsBlock = "locals"

func (l Locals) Description() string {
	return ""
}

func (l Locals) Identifier() string {
	return "locals"
}

func (d LocalsGroup) ById(id string) (*Locals, error) {
	for _, data := range d {
		attributes, err := data.Body.JustAttributes()
		if err != nil {
			return nil, err
		}
		for _, attribute := range attributes {
			if attribute.Name == id {
				return &data, nil
			}
		}

	}
	return nil, fmt.Errorf("data block with id %s not found", id)
}

func (l Locals) Type() string {
	return LocalsBlock
}

func (l Locals) IsDaemon() bool {
	return false
}

func (l Locals) Terminate() diag.Diagnostics {
	return nil
}

func (l Locals) Kill() diag.Diagnostics {
	return nil
}
