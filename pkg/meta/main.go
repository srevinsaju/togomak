package meta

import uuid "github.com/satori/go.uuid"

const (
	Version                  = "0.0.2"
	AppName                  = "togomak"
	SupportedCiConfigVersion = 1
	EnvPrefix                = "TOGOMAK"
	BuildDirPrefix           = ".togomak"
	BuildDir                 = "build"
	DefaultWorkspaceType     = "default"
	ExtendsDir               = "extends"
)

var correlationID uuid.UUID

func GetCorrelationId() uuid.UUID {
	if correlationID == uuid.Nil {
		correlationID = uuid.NewV4()
	}
	return correlationID
}
