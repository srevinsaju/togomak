package dg

import (
	"github.com/hashicorp/hcl/v2"
	"sync"
	"testing"
)

func TestSafeDiagnostics_Append(t *testing.T) {
	// test for concurrent writes to the diagnostics

	d := &SafeDiagnostics{}
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {

		wg.Add(1)
		go func() {
			d.Append(&hcl.Diagnostic{
				Severity:    0,
				Summary:     "",
				Detail:      "",
				Subject:     nil,
				Context:     nil,
				Expression:  nil,
				EvalContext: nil,
				Extra:       nil,
			})
			wg.Done()
		}()
	}
	wg.Wait()
	if len(d.Diagnostics()) != 100 {
		t.Errorf("Diagnostics not written concurrently")
	}
}

func TestSafeDiagnostics_Extend(t *testing.T) {

	d := &SafeDiagnostics{}
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {

		wg.Add(1)
		go func() {
			d.Extend(hcl.Diagnostics{
				{
					Severity:    0,
					Summary:     "",
					Detail:      "",
					Subject:     nil,
					Context:     nil,
					Expression:  nil,
					EvalContext: nil,
					Extra:       nil,
				},
			})
			wg.Done()
		}()
	}
	wg.Wait()
	if len(d.Diagnostics()) != 100 {
		t.Errorf("Diagnostics not written concurrently")
	}
}

func TestSafeDiagnostics_HasErrors(t *testing.T) {
	d := &SafeDiagnostics{}
	d.Extend(hcl.Diagnostics{
		{
			Severity:    hcl.DiagError,
			Summary:     "",
			Detail:      "",
			Subject:     nil,
			Context:     nil,
			Expression:  nil,
			EvalContext: nil,
			Extra:       nil,
		},
	})
	if !d.HasErrors() {
		t.Errorf("HasErrors not working")
	}
}

func TestSafeDiagnostics_Diagnostics(t *testing.T) {
	d := &SafeDiagnostics{}
	d.Extend(hcl.Diagnostics{
		{
			Severity:    0,
			Summary:     "",
			Detail:      "",
			Subject:     nil,
			Context:     nil,
			Expression:  nil,
			EvalContext: nil,
			Extra:       nil,
		},
	})
	if len(d.Diagnostics()) != 1 {
		t.Errorf("Diagnostics not working")
	}
}
