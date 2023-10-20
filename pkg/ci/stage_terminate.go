package ci

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	dockerContainer "github.com/docker/docker/api/types/container"
	dockerClient "github.com/docker/docker/client"
	"github.com/hashicorp/hcl/v2"
	"github.com/srevinsaju/togomak/v1/pkg/runnable"
	"syscall"
)

func (s *Stage) Terminate(conductor *Conductor, safe bool) hcl.Diagnostics {
	s.Logger().Debug("terminating stage")
	ctx := context.Background()
	var diags hcl.Diagnostics
	if safe {
		s.terminated = true
	}

	defer func() {
		diags = diags.Extend(s.AfterRun(
			conductor,
			runnable.WithHook(),
			runnable.WithStatus(runnable.StatusTerminated),
			runnable.WithParent(runnable.ParentConfig{Name: s.Name, Id: s.Id}),
		))
	}()

	if s.Container != nil && s.ContainerId != "" {

		cli, err := dockerClient.NewClientWithOpts(dockerClient.FromEnv, dockerClient.WithAPIVersionNegotiation())
		if err != nil {
			diags = diags.Append(&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "failed to create docker client",
				Detail:   fmt.Sprintf("%s: %s", dockerContainerSourceFmt(s.ContainerId), err.Error()),
			})
		}
		s.Logger().Debug("stopping container")
		err = cli.ContainerStop(ctx, s.ContainerId, dockerContainer.StopOptions{})
		if err != nil {
			diags = diags.Append(&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "failed to stop container",
				Detail:   fmt.Sprintf("%s: %s", dockerContainerSourceFmt(s.ContainerId), err.Error()),
			})
		}
		s.Logger().Debug("removing container")
		err = cli.ContainerRemove(ctx, s.ContainerId, types.ContainerRemoveOptions{
			RemoveVolumes: true,
		})
		s.Logger().Debug("removed container")

	} else if s.process != nil && s.process.Process != nil {
		if s.process.ProcessState != nil {
			if s.process.ProcessState.Exited() {
				return diags
			}
		}
		err := s.process.Process.Signal(syscall.SIGTERM)
		if err != nil {
			diags = diags.Append(&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "failed to terminate process",
				Detail:   err.Error(),
			})
		}
	}
	s.Logger().Debug("terminated stage")

	return diags
}

func (s *Stage) Kill() hcl.Diagnostics {
	diags := s.Terminate(nil, false)
	if s.process != nil && !s.process.ProcessState.Exited() {
		err := s.process.Process.Kill()
		if err != nil {
			diags = diags.Append(&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "couldn't kill stage",
				Detail:   err.Error(),
			})
		}
	}
	return diags
}

func (s *Stage) Terminated() bool {
	return s.terminated
}
