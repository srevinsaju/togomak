package runnable

import "github.com/hashicorp/hcl/v2"

type StatusType string

const (
	StatusSuccess    StatusType = "success"
	StatusFailure    StatusType = "failure"
	StatusTerminated StatusType = "terminated"
	StatusRunning    StatusType = "running"
	StatusSkipped    StatusType = "skipped"
	StatusUnknown    StatusType = "unknown"
)

func (s StatusType) String() string {
	return string(s)
}

type Status struct {
	// Diags is the diagnostics of the runnable
	Diags hcl.Diagnostics

	// Status is the status of the runnable
	Status StatusType
}
