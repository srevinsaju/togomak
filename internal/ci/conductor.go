package ci

import (
	"bytes"
	"context"
	"github.com/google/uuid"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/sirupsen/logrus"
	"github.com/srevinsaju/togomak/v1/internal/conductor"
	"github.com/srevinsaju/togomak/v1/internal/meta"
	"github.com/srevinsaju/togomak/v1/internal/x"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type ConductorOption func(*Conductor)

func ConductorWithLogger(logger logrus.Ext1FieldLogger) ConductorOption {
	return func(c *Conductor) {
		c.RootLogger = logger
	}
}

func ConductorWithConfig(cfg ConductorConfig) ConductorOption {
	return func(c *Conductor) {
		c.Config = cfg
	}
}

func ConductorWithContext(ctx context.Context) ConductorOption {
	return func(c *Conductor) {
		c.ctx = ctx
	}
}

func ConductorWithParser(parser *Parser) ConductorOption {
	return func(c *Conductor) {
		c.Parser = parser
	}
}

func ConductorWithDiagWriter(diagWriter hcl.DiagnosticWriter) ConductorOption {
	return func(c *Conductor) {
		c.DiagWriter = diagWriter
	}
}

func ConductorWithEvalContext(evalContext *hcl.EvalContext) ConductorOption {
	return func(c *Conductor) {
		c.eval.context = evalContext
	}
}

func ConductorWithProcess(process Process) ConductorOption {
	return func(c *Conductor) {
		c.Process = process
	}
}

func ConductorWithVariable(variable *Variable) ConductorOption {
	return func(c *Conductor) {
		c.variables = append(c.variables, variable)
	}
}

func ConductorWithVariablesList(variables Variables) ConductorOption {
	return func(c *Conductor) {
		c.variables = variables
	}
}

type Eval struct {
	context *hcl.EvalContext
	mu      *sync.RWMutex
}

func (e *Eval) Context() *hcl.EvalContext {
	return e.context
}

func (e *Eval) Mutex() *sync.RWMutex {
	return e.mu
}

type Conductor struct {
	RootLogger logrus.Ext1FieldLogger
	Config     ConductorConfig

	stdinMu sync.Mutex

	ctx context.Context

	// Process is the current process
	Process Process

	// Parser is the HCL parser
	Parser *Parser

	// DiagWriter is the HCL diagnostic writer, it is used to write the diagnostics
	// to os.Stdout
	DiagWriter hcl.DiagnosticWriter

	// Eval has the evaluation context and mutexes associated with Eval Variable maps
	eval *Eval

	parent *Conductor

	variables Variables

	outputsMu sync.Mutex
	outputs   map[string]*bytes.Buffer
}

func (c *Conductor) Outputs() map[string]*bytes.Buffer {
	c.outputsMu.Lock()
	defer c.outputsMu.Unlock()
	return c.outputs
}

func (c *Conductor) OutputMemoryStream(name string) *bytes.Buffer {
	c.outputsMu.Lock()
	defer c.outputsMu.Unlock()
	if c.outputs == nil {
		c.outputs = make(map[string]*bytes.Buffer)
	}
	return c.outputs[name]
}

func (c *Conductor) NewOutputMemoryStream(name string) *bytes.Buffer {
	c.outputsMu.Lock()
	defer c.outputsMu.Unlock()
	if c.outputs == nil {
		c.outputs = make(map[string]*bytes.Buffer)
	}
	c.outputs[name] = bytes.NewBuffer(nil)
	return c.outputs[name]
}

type Parser struct {
	parser *hclparse.Parser
	mu     *sync.RWMutex
}

func (p *Parser) ParseHCLFile(filename string) (*hcl.File, hcl.Diagnostics) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.parser.ParseHCLFile(filename)
}

func (p *Parser) Files() map[string]*hcl.File {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.parser.Files()
}

func (c *Conductor) TempDir() string {
	return c.Process.TempDir
}

func (c *Conductor) Eval() conductor.Eval {
	return c.eval
}

func (c *Conductor) StdinLock() {
	c.stdinMu.Lock()
}

func (c *Conductor) StdinUnlock() {
	c.stdinMu.Unlock()
}

func (c *Conductor) Child(opts ...ConductorOption) *Conductor {
	inheritOpts := []ConductorOption{
		ConductorWithConfig(c.Config),
	}
	opts = append(inheritOpts, opts...)
	child := NewConductor(c.Config, opts...)
	child.parent = c
	return child
}

func (c *Conductor) Parent() *Conductor {
	return c.parent
}

func (c *Conductor) RootParent() *Conductor {
	if c.parent == nil {
		return c
	}
	return c.parent.RootParent()
}

func (c *Conductor) Logger() logrus.Ext1FieldLogger {
	return c.RootLogger
}

type Process struct {
	Id uuid.UUID

	Executable string

	// BootTime is the time when the process was started
	BootTime time.Time

	// TempDir is the temporary directory created for the process
	TempDir string
}

func (c *Conductor) Context() context.Context {
	return c.ctx
}

func NewProcess(cfg ConductorConfig) Process {
	e, err := os.Executable()
	x.Must(err)

	pipelineId := uuid.New()

	// create a temporary directory
	tempDir, err := os.MkdirTemp("", "togomak")
	x.Must(err)

	return Process{
		Id:         pipelineId,
		Executable: e,
		BootTime:   time.Now(),
		TempDir:    tempDir,
	}
}

func (c *Conductor) Variables() Variables {
	return c.variables
}

func Chdir(cfg ConductorConfig, logger *logrus.Logger) string {
	cwd := cfg.Paths.Cwd
	if cwd == "" {
		cwd = filepath.Dir(cfg.Paths.Pipeline)
		if filepath.Base(cwd) == meta.BuildDirPrefix {
			cwd = filepath.Dir(cwd)
		}
	}
	err := os.Chdir(cwd)
	if err != nil {
		logger.Fatal(err)
	}
	cwd, err = os.Getwd()
	x.Must(err)
	logger.Debug("changing working directory to ", cwd)
	return cwd

}

func NewConductor(cfg ConductorConfig, opts ...ConductorOption) *Conductor {
	parser := hclparse.NewParser()

	diagWriter := hcl.NewDiagnosticTextWriter(os.Stdout, parser.Files(), 0, true)

	logger := NewLogger(cfg)

	dir := Chdir(cfg, logger)
	if dir != cfg.Paths.Cwd {
		cfg.Paths.Cwd = dir
	}
	if cfg.Paths.Module == "" {
		cfg.Paths.Module = cfg.Paths.Cwd
	}

	process := NewProcess(cfg)

	c := &Conductor{
		Parser: &Parser{
			parser: parser,
			mu:     &sync.RWMutex{},
		},
		DiagWriter: diagWriter,
		ctx:        context.Background(),
		Process:    process,
		RootLogger: logger,
		Config:     cfg,
		eval: &Eval{
			context: CreateEvalContext(cfg, process),
			mu:      &sync.RWMutex{},
		},
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

func (c *Conductor) Update(opts ...ConductorOption) {
	for _, opt := range opts {
		opt(c)
	}
}
