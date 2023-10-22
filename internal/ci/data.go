package ci

import (
	"github.com/hashicorp/hcl/v2"
)

const DataBlock = "data"

func (s *Data) Description() Description {
	return Description{Name: s.Name}
}

func (s *Data) Identifier() string {
	return s.Id
}

func (d Datas) ById(provider string, id string) (*Data, hcl.Diagnostics) {
	for _, data := range d {
		if data.Id == id && data.Provider == provider {
			return &data, nil
		}
	}
	return nil, hcl.Diagnostics{
		{
			Severity: hcl.DiagError,
			Summary:  "Data not found",
			Detail:   "Data with id " + id + " not found",
		},
	}
}

func (s *Data) Type() string {
	return DataBlock
}

func (s *Data) Set(k any, v any) {
}

func (s *Data) Get(k any) any {
	return nil
}

func (s *Data) IsDaemon() bool {
	return false
}

func (s *Data) Terminate(conductor *Conductor, safe bool) hcl.Diagnostics {
	return nil
}

func (s *Data) Kill() hcl.Diagnostics {
	return nil
}
