package context

import "os/exec"

// RunningStage schema will be used to store references to running stages
// and can be used to terminate running stages later
type RunningStage struct {
	Id      string
	Process *exec.Cmd
}
