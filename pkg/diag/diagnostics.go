package diag

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/srevinsaju/togomak/v1/pkg/ui"
	"io"
	"os"
)

const (
	SeverityError   = "error"
	SeverityWarning = "warning"
	SeverityInvalid = "invalid"
)

type Diagnostic struct {
	Severity string
	Summary  string
	Detail   string
	Source   string
}

func NewDiagnostic(severity, summary, detail, source string) Diagnostic {
	return Diagnostic{
		Severity: severity,
		Summary:  summary,
		Detail:   detail,
		Source:   source,
	}
}

type Diagnostics []Diagnostic

func (d Diagnostics) Len() int {
	return len(d)
}

func (d Diagnostics) Extend(diags Diagnostics) Diagnostics {
	return append(d, diags...)
}

func (d Diagnostics) ExtendHCLDiagnostics(hclDiags hcl.Diagnostics, hclDgWriter hcl.DiagnosticWriter, source string) Diagnostics {
	if hclDiags.HasErrors() {
		err := hclDgWriter.WriteDiagnostics(hclDiags)
		if err != nil {
			d = d.Append(Diagnostic{
				Severity: SeverityError,
				Summary:  "failed to write HCL diagnostics",
				Detail:   err.Error(),
			})
		}
		d = d.Append(Diagnostic{
			Severity: SeverityError,
			Summary:  "failed to evaluate HCL",
			Detail:   hclDiags.Error(),
			Source:   source,
		})
	}

	return d
}

func (d Diagnostics) Append(diag Diagnostic) Diagnostics {
	return append(d, diag)
}

func (d Diagnostics) HasErrors() bool {
	for _, diag := range d {
		if diag.Severity == "error" {
			return true
		}
	}
	return false
}
func (d Diagnostic) Error() string {
	message := fmt.Sprintf("%s: %s, (%s)", d.Severity, d.Summary, d.Detail)
	return message
}

func (d Diagnostics) Error() string {
	count := len(d)
	switch {
	case count == 0:
		return "no diagnostics"
	case count == 1:
		return d[0].Error()
	default:
		return fmt.Sprintf("%s, and %d other diagnostic(s)", d[0].Error(), count-1)
	}
}

func (d Diagnostics) HasWarnings() bool {
	for _, diag := range d {
		if diag.Severity == SeverityWarning {
			return true
		}
	}
	return false
}

func NewNotImplementedError(source string) Diagnostic {
	return NewError(source, "Runtime Error", "not implemented")
}

func NewError(source, summary, message string) Diagnostic {
	return NewDiagnostic(SeverityError, summary, message, source)
}

func (d Diagnostics) NewHclWriteDiagnosticsError(source string, err error) Diagnostics {
	diags := d
	if err != nil {
		diags = diags.Append(Diagnostic{
			Severity: SeverityError,
			Summary:  "failed to write diagnostics",
			Detail:   err.Error(),
			Source:   source,
		})
	}
	diags = diags.Append(Diagnostic{
		Severity: SeverityError,
		Summary:  "failed to decode body",
		Detail:   err.Error(),
		Source:   source,
	})
	return diags
}

func (d Diagnostics) Write(writer io.Writer) {
	for i, diag := range d {

		diagHeader := ui.Red(fmt.Sprintf("Error %d", i+1))
		if diag.Severity == SeverityWarning {
			diagHeader = ui.Yellow(fmt.Sprintf("Warning %d", i+1))
		}
		_, err := fmt.Fprintf(writer, "%s: %s\n\t%s\n\tsource: %s\n\n",
			diagHeader,
			ui.Bold(diag.Summary),
			diag.Detail,
			ui.Grey(diag.Source),
		)
		if err != nil {
			panic(err)
		}
	}
	writer.Write([]byte("\n"))
}

func (d Diagnostics) Fatal(writer io.Writer) {
	d.Write(writer)
	os.Exit(1)
}
