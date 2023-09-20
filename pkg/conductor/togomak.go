package conductor

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/sirupsen/logrus"
	"os"
)

type Togomak struct {
	Logger *logrus.Logger
	Config Config

	// hcl stuff

	// Parser is the HCL parser
	Parser *hclparse.Parser

	// DiagWriter is the HCL diagnostic writer, it is used to write the diagnostics
	// to os.Stdout
	DiagWriter hcl.DiagnosticWriter
}

func NewTogomak(cfg Config) *Togomak {
	parser := hclparse.NewParser()
	diagWriter := hcl.NewDiagnosticTextWriter(os.Stdout, parser.Files(), 0, true)

	return &Togomak{
		Parser:     parser,
		DiagWriter: diagWriter,

		Logger: NewLogger(cfg),
		Config: cfg,
	}
}
