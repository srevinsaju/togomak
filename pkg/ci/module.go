package ci

import "github.com/hashicorp/hcl/v2"

const ModuleBlock = "module"

func (m *Module) Description() string {
	return ""
}

func (m *Module) Identifier() string {
	if m.Id == "" {
		panic("id not set")
	}
	return m.Id
}

func (i Modules) ById(id string) (*Module, hcl.Diagnostics) {
	for _, macro := range i {
		if macro.Identifier() == id {
			return &macro, nil
		}
	}
	return nil, hcl.Diagnostics{
		{
			Severity: hcl.DiagError,
			Summary:  "module not found",
			Detail:   "module with id " + id + " not found",
		},
	}
}

func (m *Module) Type() string {
	return ModuleBlock
}

func (m *Module) IsDaemon() bool {
	return false
}

func (m *Module) Terminate(safe bool) hcl.Diagnostics {
	return nil
}

func (m *Module) Kill() hcl.Diagnostics {
	return nil
}

func (m *Module) Set(k any, v any) {
}

func (m *Module) Get(k any) any {
	return nil
}
