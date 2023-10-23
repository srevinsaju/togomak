package ci

import (
	"github.com/google/uuid"
	"github.com/srevinsaju/togomak/v1/internal/global"
	"github.com/srevinsaju/togomak/v1/internal/x"
	"os"
	"time"
)

type Process struct {
	Id uuid.UUID

	Executable string

	// BootTime is the time when the process was started
	BootTime time.Time

	// TempDir is the temporary directory created for the process
	TempDir string
}

func NewProcess(cfg ConductorConfig) Process {
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
