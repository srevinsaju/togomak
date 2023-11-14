package ci

import (
	"context"
	"github.com/hashicorp/hcl/v2"
	"github.com/srevinsaju/togomak/v1/internal/c"
	"github.com/srevinsaju/togomak/v1/internal/dg"
	"github.com/srevinsaju/togomak/v1/internal/runnable"
)

func StartHandlers(conductor *Conductor) *Handler {

	h := NewHandler(
		WithContext(conductor.Context()),
		WithLogger(conductor.RootLogger),
		WithDiagnosticWriter(conductor.DiagWriter),
		WithProcessBootTime(conductor.Process.BootTime),
	)
	go h.Interrupt()
	go h.Kill()
	go h.Daemons()
	return h
}

func (pipe *Pipeline) Run(conductor *Conductor) (*Handler, dg.AbstractDiagnostics) {
	var d hcl.Diagnostics
	logger := conductor.Logger().WithField("orchestra", "run")
	cfg := conductor.Config
	ctx, cancel := context.WithCancel(conductor.Context())
	h := StartHandlers(conductor)

	defer cancel()
	defer h.WriteDiagnostics()

	// --> expand imports
	pipe, d = ExpandImports(conductor, pipe, conductor.Config.Paths)
	h.Diags.Extend(d)
	if h.Diags.HasErrors() {
		return h, h.Diags
	}

	/// we will first expand all local blocks
	logger.Debugf("expanding local blocks")
	locals, d := pipe.Locals.Expand()
	h.Diags.Extend(d)
	if d.HasErrors() {
		return h, h.Diags
	}
	pipe.Local = locals

	// store the pipe in the context
	ctx = context.WithValue(ctx, c.TogomakContextPipeline, pipe)
	h = h.Update(WithContext(ctx))
	conductor.Update(ConductorWithContext(ctx))

	// --> generate a dependency graph
	// we will now generate a dependency graph from the pipeline
	// this will be used to generate the pipeline
	logger.Debugf("generating dependency graph")
	depGraph, d := GraphTopoSort(conductor, pipe)
	h.Diags.Extend(d)
	if h.Diags.HasErrors() {
		return h, h.Diags
	}

	// endregion: interrupt h
	opts := []runnable.Option{
		runnable.WithBehavior(conductor.Config.Behavior),
		runnable.WithPaths(conductor.Config.Paths),
	}

	logger.Debugf("starting runnables")
	for _, layer := range depGraph.TopoSortedLayers() {
		// we parse the TOGOMAK_ENV file at the beginning of every layer
		// this allows us to have different environments for different layers

		d = ExpandOutputs(conductor)
		h.Diags.Extend(d)
		if h.Diags.HasErrors() {
			break
		}

		for _, runnableId := range layer {

			runnable, skip, d := pipe.Resolve(runnableId)
			if skip {
				continue
			}
			if d.HasErrors() {
				h.Diags.Extend(d)
				break
			}

			ok, overridden, d := BlockCanRun(runnable, conductor, runnableId, depGraph, opts...)
			h.Diags.Extend(d)
			if d.HasErrors() {
				break
			}

			// prepare step needs to pipeline.Run before the runnable is pipeline.Run
			// we will also need to prompt the user with the information saying that it has been skipped
			d = runnable.Prepare(conductor, !ok, overridden)
			h.Diags.Extend(d)
			if d.HasErrors() {
				break
			}

			if !ok {
				logger.Debugf("skipping runnable %s, condition evaluated to false", runnableId)
				continue
			}

			logger.Debugf("runnable %s is %T", runnableId, runnable)

			if runnable.IsDaemon() {
				h.Tracker.AppendDaemon(runnable)
			} else {
				h.Tracker.AppendRunnable(runnable)
			}

			go BlockRunWithRetries(conductor, runnableId, runnable, h, conductor.Logger(), opts...)

			if cfg.Pipeline.DryRun {
				// TODO: implement --concurrency option
				// wait for the runnable to finish
				// disable concurrency
				h.Tracker.RunnableWait()
				h.Tracker.DaemonWait()
			}
			if pipe.Builder.Behavior != nil && pipe.Builder.Behavior.DisableConcurrency {
				h.Tracker.RunnableWait()
				h.Tracker.DaemonWait()
			}
		}
		h.Tracker.RunnableWait()

		if h.Diags.HasErrors() {
			if h.Tracker.HasDaemons() && !cfg.Pipeline.DryRun && !cfg.Behavior.Unattended {
				logger.Info("pipeline failed, waiting for daemons to shut down")
				logger.Info("hit Ctrl+C to force stop them")
				// wait for daemons to stop
				h.Tracker.DaemonWait()
			} else if h.Tracker.HasDaemons() && !cfg.Pipeline.DryRun {
				logger.Info("pipeline failed, waiting for daemons to shut down...")
				// wait for daemons to stop
				cancel()
			}
			break
		}
	}

	h.Tracker.DaemonWait()
	return h, h.Diags
}
