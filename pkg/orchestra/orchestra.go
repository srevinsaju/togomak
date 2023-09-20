package orchestra

import (
	"context"
	"github.com/hashicorp/hcl/v2"
	"github.com/srevinsaju/togomak/v1/pkg/c"
	"github.com/srevinsaju/togomak/v1/pkg/ci"
	"github.com/srevinsaju/togomak/v1/pkg/conductor"
	"github.com/srevinsaju/togomak/v1/pkg/global"
	"github.com/srevinsaju/togomak/v1/pkg/graph"
	"github.com/srevinsaju/togomak/v1/pkg/meta"
	"github.com/srevinsaju/togomak/v1/pkg/pipeline"
	"strings"

	"github.com/zclconf/go-cty/cty"
	"os"
	"path/filepath"
	"sync"
)

func loadGlobalParams(t *Togomak, cfg conductor.Config) {
	paramsGo := make(map[string]cty.Value)
	if cfg.Behavior.Child.Enabled {
		m := make(map[string]string)
		for _, e := range os.Environ() {
			if i := strings.Index(e, "="); i >= 0 {
				if strings.HasPrefix(e[:i], ci.TogomakParamEnvVarPrefix) {
					m[e[:i]] = e[i+1:]
				}
			}
		}
		for k, v := range m {
			if ci.TogomakParamEnvVarRegex.MatchString(k) {
				paramsGo[ci.TogomakParamEnvVarRegex.FindStringSubmatch(k)[1]] = cty.StringVal(v)
			}
		}
	}
	global.EvalContextMutex.Lock()
	t.ectx.Variables[ci.ParamBlock] = cty.ObjectVal(paramsGo)
	global.EvalContextMutex.Unlock()
}

func Perform(cfg conductor.Config) int {

	t, ctx := NewContextWithTogomak(cfg)
	ctx, cancel := context.WithCancel(ctx)

	logger := t.Logger
	logger.Debugf("starting watchdogs and signal handlers")
	handler := StartHandlers(ctx)

	defer cancel()
	defer handler.WriteDiagnostics(&t)

	// region: external parameters
	loadGlobalParams(&t, cfg)
	// endregion

	// --> parse the config file
	// we will now read the pipeline from togomak.hcl
	pipe, hclDiags := pipeline.Read(ctx, t.parser)
	if hclDiags.HasErrors() {
		logger.Fatal(t.hclDiagWriter.WriteDiagnostics(hclDiags))
	}

	// whitelist all stages if unspecified
	filterList := cfg.Pipeline.Filtered

	// write the pipeline to the temporary directory
	pipelineFilePath := filepath.Join(t.cwd, t.tempDir, meta.ConfigFileName)
	var pipelineData []byte
	for _, f := range t.parser.Files() {
		pipelineData = append(pipelineData, f.Bytes...)
	}

	err := os.WriteFile(pipelineFilePath, pipelineData, 0644)
	if err != nil {
		return fatal(ctx)
	}
	var d hcl.Diagnostics

	pipe, d = ExpandImports(pipe, ctx, t)
	handler.Diags.Extend(d)
	if handler.Diags.HasErrors() {
		return fatal(ctx)
	}

	/// we will first expand all local blocks
	logger.Debugf("expanding local blocks")
	locals, d := pipe.Locals.Expand()
	handler.Diags.Extend(d)
	if d.HasErrors() {
		return fatal(ctx)
	}
	pipe.Local = locals

	// store the pipe in the context
	ctx = context.WithValue(ctx, c.TogomakContextPipeline, pipe)

	// --> validate the pipeline
	// TODO: validate the pipeline

	// --> generate a dependency graph
	// we will now generate a dependency graph from the pipeline
	// this will be used to generate the pipeline
	logger.Debugf("generating dependency graph")
	depGraph, d := graph.TopoSort(ctx, pipe)
	handler.Diags.Extend(d)
	if handler.Diags.HasErrors() {
		return fatal(ctx)
	}

	// endregion: interrupt handler

	var diagsMutex sync.Mutex

	logger.Debugf("starting runnables")
	for _, layer := range depGraph.TopoSortedLayers() {
		// we parse the TOGOMAK_ENV file at the beginning of every layer
		// this allows us to have different environments for different layers

		d = ExpandOutputs(t, logger)
		handler.Diags.Extend(d)
		if handler.Diags.HasErrors() {
			break
		}

		for _, runnableId := range layer {

			runnable, skip, d := pipe.Resolve(runnableId)
			if skip {
				continue
			}
			if d.HasErrors() {
				diagsMutex.Lock()
				handler.Diags.Extend(d)
				diagsMutex.Unlock()
				break
			}

			ok, d, overridden := CanRun(runnable, ctx, filterList, runnableId, depGraph)
			diagsMutex.Lock()
			handler.Diags.Extend(d)
			diagsMutex.Unlock()
			if d.HasErrors() {
				break
			}

			// prepare step needs to run before the runnable is run
			// we will also need to prompt the user with the information saying that it has been skipped
			d = runnable.Prepare(ctx, !ok, overridden)
			diagsMutex.Lock()
			handler.Diags.Extend(d)
			diagsMutex.Unlock()
			if d.HasErrors() {
				break
			}

			if !ok {
				logger.Debugf("skipping runnable %s, condition evaluated to false", runnableId)
				continue
			}

			logger.Debugf("runnable %s is %T", runnableId, runnable)

			if runnable.IsDaemon() {
				handler.Tracker.AppendDaemon(runnable)
			} else {
				handler.Tracker.AppendRunnable(runnable)
			}

			go RunWithRetries(runnableId, runnable, ctx, handler, logger)

			if cfg.Pipeline.DryRun {
				// TODO: implement --concurrency option
				// wait for the runnable to finish
				// disable concurrency
				handler.Tracker.RunnableWait()
				handler.Tracker.DaemonWait()
			}
		}
		handler.Tracker.RunnableWait()

		if handler.Diags.HasErrors() {
			if handler.Tracker.HasDaemons() && !cfg.Pipeline.DryRun && !cfg.Behavior.Unattended {
				logger.Info("pipeline failed, waiting for daemons to shut down")
				logger.Info("hit Ctrl+C to force stop them")
				// wait for daemons to stop
				handler.Tracker.DaemonWait()
			} else if handler.Tracker.HasDaemons() && !cfg.Pipeline.DryRun {
				logger.Info("pipeline failed, waiting for daemons to shut down...")
				// wait for daemons to stop
				cancel()
			}
			break
		}
	}

	handler.Tracker.DaemonWait()
	if handler.Diags.HasErrors() {
		return fatal(ctx)
	}
	return ok(ctx)
}

func StartHandlers(ctx context.Context) *Handler {
	handler := NewHandler(ctx)
	go handler.Interrupt()
	go handler.Kill()
	go handler.Daemons()
	return handler
}
