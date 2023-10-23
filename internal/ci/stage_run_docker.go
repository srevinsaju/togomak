package ci

import (
	"context"
	"errors"
	"fmt"
	"github.com/docker/docker/api/types"
	dockerContainer "github.com/docker/docker/api/types/container"
	dockerClient "github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/hashicorp/hcl/v2"
	"github.com/sirupsen/logrus"
	"github.com/srevinsaju/togomak/v1/internal/ui"
	"github.com/srevinsaju/togomak/v1/internal/x"
	"io"
)

func (s *stageConductor) runUsingDocker() *stageConductor {
	var diags hcl.Diagnostics
	logger := s.Logger()
	ctx := s.conductor.Context()

	image, d := s.hclImage(s.conductor, s.Eval().Context())
	diags = diags.Extend(d)

	// begin entrypoint evaluation
	entrypoint, d := s.hclEndpoint(s.conductor, s.Eval().Context())
	diags = diags.Extend(d)

	if diags.HasErrors() {
		s.diags = s.diags.Extend(diags)
		return s
	}

	cli, err := dockerClient.NewClientWithOpts(dockerClient.FromEnv, dockerClient.WithAPIVersionNegotiation())
	if err != nil {
		diags = diags.Append(&hcl.Diagnostic{
			Severity:    hcl.DiagError,
			Summary:     "could not create docker client",
			Detail:      err.Error(),
			Subject:     s.Container.Image.Range().Ptr(),
			EvalContext: s.Eval().Context(),
		})
		s.diags = s.diags.Extend(diags)
		return s
	}
	defer x.Must(cli.Close())

	// check if image exists
	logger.Debugf("checking if image %s exists", image)
	_, _, err = cli.ImageInspectWithRaw(ctx, image)
	if err != nil {
		logger.Infof("image %s does not exist, pulling...", image)
		reader, err := cli.ImagePull(ctx, image, types.ImagePullOptions{})
		if err != nil {
			diags = diags.Append(&hcl.Diagnostic{
				Severity:    hcl.DiagError,
				Summary:     "could not pull image",
				Detail:      err.Error(),
				Subject:     s.Container.Image.Range().Ptr(),
				EvalContext: s.Eval().Context(),
			})
			s.diags = s.diags.Extend(diags)
			return s
		}

		pb := ui.NewDockerProgressWriter(reader, logger.Writer(), fmt.Sprintf("pulling image %s", image))
		defer pb.Close()
		defer reader.Close()
		io.Copy(pb, reader)
	}

	logger.Trace("parsing container arguments")
	binds := []string{
		fmt.Sprintf("%s:/workspace", s.attr.cmd.Dir),
	}

	logger.Trace("parsing container volumes")
	for _, m := range s.Container.Volumes {
		s.Eval().Mutex().RLock()
		source, d := m.Source.Value(s.Eval().Context())
		s.Eval().Mutex().RUnlock()
		diags = diags.Extend(d)

		s.Eval().Mutex().RLock()
		dest, d := m.Destination.Value(s.Eval().Context())
		s.Eval().Mutex().RUnlock()
		diags = diags.Extend(d)
		if diags.HasErrors() {
			continue
		}
		binds = append(binds, fmt.Sprintf("%s:%s", source.AsString(), dest.AsString()))
	}
	logger.Tracef("%d diagnostic(s) after parsing container volumes", len(diags.Errs()))
	if diags.HasErrors() {
		s.diags = s.diags.Extend(diags)
		return s
	}

	logger.Trace("dry run check")
	if s.cfg.Behavior.DryRun {
		fmt.Println(ui.Blue("docker:run.image"), ui.Green(image))
		fmt.Println(ui.Blue("docker:run.workdir"), ui.Green("/workspace"))
		fmt.Println(ui.Blue("docker:run.volume"), ui.Green(s.attr.cmd.Dir+":/workspace"))
		fmt.Println(ui.Blue("docker:run.stdin"), ui.Green(s.Container.Stdin))
		fmt.Println(ui.Blue("docker:run.args"), ui.Green(s.attr.cmd.String()))
		s.diags = s.diags.Extend(diags)
		return s
	}

	logger.Trace("parsing container ports")
	exposedPorts, bindings, d := s.Container.Ports.Nat(s.conductor, s.Eval().Context())
	diags = diags.Extend(d)
	if diags.HasErrors() {
		s.diags = s.diags.Extend(diags)
		return s
	}

	logger.Trace("creating container")
	resp, err := cli.ContainerCreate(s.conductor.Context(), &dockerContainer.Config{
		Image:      image,
		Cmd:        s.attr.cmd.Args,
		WorkingDir: "/workspace",
		Volumes: map[string]struct{}{
			"/workspace": {},
		},
		Tty:          true,
		AttachStdout: true,
		AttachStderr: true,
		AttachStdin:  s.Container.Stdin,
		OpenStdin:    s.Container.Stdin,
		StdinOnce:    s.Container.Stdin,
		Entrypoint:   entrypoint,
		Env:          s.attr.cmd.Env,
		ExposedPorts: exposedPorts,
		// User: s.Container.User,
	}, &dockerContainer.HostConfig{
		Binds:        binds,
		PortBindings: bindings,
	}, nil, nil, "")
	if err != nil {
		diags = diags.Append(&hcl.Diagnostic{
			Severity:    hcl.DiagError,
			Summary:     "could not create container",
			Detail:      err.Error(),
			Subject:     s.Container.Image.Range().Ptr(),
			EvalContext: s.Eval().Context(),
		})
		s.diags = s.diags.Extend(diags)
		return s
	}

	logger.Trace("starting container")
	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		diags = diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "could not start container",
			Detail:   err.Error(),
			Subject:  s.Container.Image.Range().Ptr(),
		})
		s.diags = s.diags.Extend(diags)
		return s
	}
	s.ContainerId = resp.ID

	logger.Trace("getting container metadata for log retrieval")
	container, err := cli.ContainerInspect(ctx, resp.ID)
	if err != nil {
		panic(err)
	}

	logger.Trace("getting container logs")
	responseBody, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{
		ShowStdout: true, ShowStderr: true,
		Follow: true,
	})
	if err != nil {
		diags = diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "could not get container logs",
			Detail:   err.Error(),
			Subject:  s.Container.Image.Range().Ptr(),
		})
		s.diags = s.diags.Extend(diags)
		return s
	}
	defer responseBody.Close()

	logger.Tracef("copying container logs on container: %s", resp.ID)
	if container.Config.Tty {
		_, err = io.Copy(logger.Writer(), responseBody)
	} else {
		_, err = stdcopy.StdCopy(logger.Writer(), logger.WriterLevel(logrus.WarnLevel), responseBody)
	}

	logger.Trace("waiting for container to finish")
	if err != nil && err != io.EOF {
		if errors.Is(err, context.Canceled) {
			s.diags = s.diags.Extend(diags)
			return s
		}
		diags = diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "failed to copy container logs",
			Detail:   err.Error(),
			Subject:  s.Container.Image.Range().Ptr(),
		})
	}

	logger.Tracef("removing container with id: %s", resp.ID)
	err = cli.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{
		RemoveVolumes: true,
	})
	if err != nil {
		diags = diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "failed to remove container",
			Detail:   err.Error(),
			Subject:  s.Container.Image.Range().Ptr(),
		})
		s.diags = s.diags.Extend(diags)
		return s
	}

	logger.Tracef("%d diagnostic(s) after removing container", len(diags.Errs()))
	s.diags = s.diags.Extend(diags)
	return s
}
