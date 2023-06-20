package ci

import (
	"context"
	"fmt"
	"github.com/srevinsaju/togomak/v1/pkg/diag"
)

const DataBlock = "data"

func (d Data) Description() string {
	return ""
}

func (d Data) Identifier() string {
	return d.Id
}

func (d Datas) ById(provider string, id string) (*Data, error) {
	for _, data := range d {
		if data.Id == id && data.Provider == provider {
			return &data, nil
		}
	}
	return nil, fmt.Errorf("data block with id %s not found", id)
}

func (d Data) Type() string {
	return DataBlock
}

func (d Data) IsDaemon() bool {
	return false
}

func (d Data) Terminate() diag.Diagnostics {
	return nil
}

func (d Data) Kill() diag.Diagnostics {
	return nil
}

func (d Data) Expand(ctx context.Context) (Runnables, diag.Diagnostics) {
	return nil, nil
}
func (d Data) Expanded() bool {
	return false
}
func (d Data) ForEachDerived() bool {
	return false
}
