package conductor

import (
	"context"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/sirupsen/logrus"
	"github.com/srevinsaju/togomak/v1/pkg/x"
	"os"
	"time"
)

type Togomak struct {
	Logger  *logrus.Logger
	Config  Config
	Context context.Context

	// Process is the current process
	Process Process

	// hcl stuff

	// Parser is the HCL parser
	Parser *hclparse.Parser

	// DiagWriter is the HCL diagnostic writer, it is used to write the diagnostics
	// to os.Stdout
	DiagWriter hcl.DiagnosticWriter
}

type Process struct {
	Executable string

	// BootTime is the time when the process was started
	BootTime time.Time
}

func NewProcess(cfg Config) Process {
	e, err := os.Executable()
	x.Must(err)

	return Process{
		Executable: e,
		BootTime:   time.Now(),
	}
}

func NewTogomak(cfg Config) *Togomak {
	parser := hclparse.NewParser()
	diagWriter := hcl.NewDiagnosticTextWriter(os.Stdout, parser.Files(), 0, true)

	return &Togomak{
		Parser:     parser,
		DiagWriter: diagWriter,
		Context:    context.Background(),

		Process: NewProcess(cfg),

		Logger: NewLogger(cfg),
		Config: cfg,
	}
}
