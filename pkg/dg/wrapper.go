package dg

import (
	"github.com/hashicorp/hcl/v2"
	"sync"
)

type AbstractDiagnostics interface {
	Append(diag *hcl.Diagnostic)
	Extend(diags hcl.Diagnostics)
	HasErrors() bool
	Diagnostics() hcl.Diagnostics
	Error() string
	Errs() []error

	Safe() *SafeDiagnostics
	Unsafe() *Diagnostics
}

type SafeDiagnostics struct {
	diagsMu sync.Mutex
	diags   hcl.Diagnostics

	expired bool
}

func (d *SafeDiagnostics) Append(diag *hcl.Diagnostic) {
	if d.expired {
		panic("Diagnostics expired")
	}
	d.diagsMu.Lock()
	defer d.diagsMu.Unlock()
	d.diags = d.diags.Append(diag)
}

func (d *SafeDiagnostics) Extend(diags hcl.Diagnostics) {
	if d.expired {
		panic("Diagnostics expired")
	}
	d.diagsMu.Lock()
	defer d.diagsMu.Unlock()
	d.diags = d.diags.Extend(diags)
}

func (d *SafeDiagnostics) HasErrors() bool {
	return d.diags.HasErrors()
}

func (d *SafeDiagnostics) Diagnostics() hcl.Diagnostics {
	return d.diags
}

func (d *SafeDiagnostics) Error() string {
	return d.diags.Error()
}

func (d *SafeDiagnostics) Errs() []error {
	return d.diags.Errs()
}

func (d *SafeDiagnostics) Unsafe() *Diagnostics {
	d.expired = true
	return &Diagnostics{
		diags: d.diags,
	}
}

func (d *SafeDiagnostics) Safe() *SafeDiagnostics {
	return d
}

type Diagnostics struct {
	expired bool
	diags   hcl.Diagnostics
}

func (d *Diagnostics) Append(diag *hcl.Diagnostic) {
	if d.expired {
		panic("Diagnostics expired")
	}
	d.diags = d.diags.Append(diag)
}

func (d *Diagnostics) Extend(diags hcl.Diagnostics) {
	if d.expired {
		panic("Diagnostics expired")
	}
	d.diags = d.diags.Extend(diags)
}

func (d *Diagnostics) HasErrors() bool {
	return d.diags.HasErrors()
}

func (d *Diagnostics) Diagnostics() hcl.Diagnostics {
	return d.diags
}

func (d *Diagnostics) Error() string {
	return d.diags.Error()
}

func (d *Diagnostics) Errs() []error {
	return d.diags.Errs()
}

func (d *Diagnostics) Safe() *SafeDiagnostics {
	d.expired = true
	return &SafeDiagnostics{
		diags: d.diags,
	}
}

func (d *Diagnostics) Unsafe() *Diagnostics {
	return d
}
