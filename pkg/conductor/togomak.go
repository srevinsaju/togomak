package conductor

import (
	"context"
	"github.com/google/uuid"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/sirupsen/logrus"
	"github.com/srevinsaju/togomak/v1/pkg/global"
	"github.com/srevinsaju/togomak/v1/pkg/meta"
	"github.com/srevinsaju/togomak/v1/pkg/x"
	"os"
	"path/filepath"
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

	// EvalContext is the HCL evaluation context
	EvalContext *hcl.EvalContext
}

type Process struct {
	Id uuid.UUID

	Executable string

	// BootTime is the time when the process was started
	BootTime time.Time

	// TempDir is the temporary directory created for the process
	TempDir string
}

func NewProcess(cfg Config) Process {
	e, err := os.Executable()
	x.Must(err)

	pipelineId := uuid.New()

	// create a temporary directory
	tempDir, err := os.MkdirTemp("", "togomak")
	x.Must(err)
	global.SetTempDir(tempDir)

	return Process{
		Id:         pipelineId,
		Executable: e,
		BootTime:   time.Now(),
		TempDir:    tempDir,
	}
}

func Chdir(cfg Config, logger *logrus.Logger) string {
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

func NewTogomak(cfg Config) *Togomak {
	parser := hclparse.NewParser()

	diagWriter := hcl.NewDiagnosticTextWriter(os.Stdout, parser.Files(), 0, true)

	logger := NewLogger(cfg)
	global.SetLogger(logger)

	cfg.Paths.Cwd = Chdir(cfg, logger)

	if !cfg.Behavior.Child.Enabled {
		logger.Infof("%s (version=%s)", meta.AppName, meta.AppVersion)
	}

	process := NewProcess(cfg)

	return &Togomak{
		Parser:     parser,
		DiagWriter: diagWriter,
		Context:    context.Background(),

		Process: process,

		Logger: logger,
		Config: cfg,

		EvalContext: CreateEvalContext(cfg, process),
	}
}

func (t *Togomak) Destroy() {
	t.Logger.Debug("removing temporary directory")
	err := os.RemoveAll(t.Process.TempDir)
	if err != nil {
		t.Logger.Warnf("failed to remove temporary directory: %s", err)
	}

	t.Logger.Debug("destroying togomak")

	t.Logger = nil
	t.Config = Config{}
	t.Context = nil
	t.Parser = nil
	t.DiagWriter = nil
}
