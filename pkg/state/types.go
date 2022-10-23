package state

import (
	uuid "github.com/satori/go.uuid"
	"os"
	"time"
)

const Version = 1
const WorkspaceDataKey = "state_workspace"

type State struct {
	LastModified              time.Time `json:"lastModified"`
	Version                   int       `json:"version"`
	LastUsernameHumanReadable string    `json:"user"`
	LastUsernameUUID          uuid.UUID `json:"userUUID"`

	StageId string `json:"stage_id"`
}

func (s State) IsStateVersionCompatible() bool {
	return s.Version == Version
}

func (s State) IsTargetUpToDate(file string) bool {
	fileInfo, err := os.Stat(file)
	if err != nil {
		return false
	}
	v := fileInfo.ModTime().Before(s.LastModified)
	return v
}
