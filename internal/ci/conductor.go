package ci

import (
	"context"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/sirupsen/logrus"
	"github.com/srevinsaju/togomak/v1/internal/global"
	"github.com/srevinsaju/togomak/v1/internal/meta"
	"os"
	"sync"
)

// Conductor is a struct which holds the state of the orchestra
// a conductor orchestrates the orchestra, in this case, workflow or the pipeline
// the conductor is responsible for the following:
// 1. Create a temporary directory, Config.Paths.TempDir
// 2. Create a process, Process
// 3. Create a logger, Logger
// 4. Create a parser, Parser
// 5. Create a diagnostic writer DiagWriter
// 6. Create an evaluation context Eval
// 7. Create a context, Context
// 8. Create an input variable map, Variables
type Conductor struct {
	RootLogger logrus.Ext1FieldLogger
	Config     ConductorConfig

	stdinMu sync.Mutex

	ctx context.Context

	// Process is the current process
	Process Process

	// Parser is the HCL parser
	Parser *hclparse.Parser

	// DiagWriter is the HCL diagnostic writer, it is used to write the diagnostics
	// to os.Stdout
	DiagWriter hcl.DiagnosticWriter

	// Eval has the evaluation context and mutexes associated with Eval Variable maps
	eval *Eval

	parent *Conductor

	variables Variables
}

// NewConductor creates a new conductor, it accepts a ConductorConfig along with a set of options
// which can be used to override the default values
// ConductorConfig is created from command-line context using cli.Context. Extra options can be
// passed using ConductorOption
func NewConductor(cfg ConductorConfig, opts ...ConductorOption) *Conductor {
	parser := hclparse.NewParser()

	diagWriter := hcl.NewDiagnosticTextWriter(os.Stdout, parser.Files(), 0, true)

	logger := NewLogger(cfg)
	global.SetLogger(logger)

	dir := chdir(cfg, logger)
	cfg.Paths.Cwd = dir

	if cfg.Paths.Module == "" {
		cfg.Paths.Module = cfg.Paths.Cwd
	}

	process := NewProcess(cfg)

	c := &Conductor{
		Parser:     parser,
		DiagWriter: diagWriter,
		ctx:        context.Background(),
		Process:    process,
		RootLogger: logger,
		Config:     cfg,
		eval:       NewEval(cfg, process),
	}
	for _, v := range cfg.Variables {
		c.variables = append(c.variables, v)
	}

	for _, opt := range opts {
		opt(c)
	}

	if !c.Config.Behavior.Child.Enabled {
		logger.Infof("%s (version=%s)", meta.AppName, meta.AppVersion)
	}

	return c
}

// Destroy destroys the conductor, it removes the temporary directory
func (c *Conductor) Destroy() {
	c.Logger().Debug("removing temporary directory")
	err := os.RemoveAll(c.Process.TempDir)
	if err != nil {
		c.Logger().Warnf("failed to remove temporary directory: %s", err)
	}

	c.Logger().Debug("destroying togomak")

	c.RootLogger = nil
	c.Config = ConductorConfig{}
	c.Parser = nil
	c.DiagWriter = nil
}

// Update updates the conductor with the given options
func (c *Conductor) Update(opts ...ConductorOption) {
	for _, opt := range opts {
		opt(c)
	}
}
