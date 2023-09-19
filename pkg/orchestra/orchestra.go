package orchestra

import (
	"context"
	"github.com/hashicorp/go-envparse"
	"github.com/hashicorp/hcl/v2"
	"github.com/kendru/darwin/go/depgraph"
	"github.com/srevinsaju/togomak/v1/pkg/c"
	"github.com/srevinsaju/togomak/v1/pkg/ci"
	"github.com/srevinsaju/togomak/v1/pkg/filter"
	"github.com/srevinsaju/togomak/v1/pkg/global"
	"github.com/srevinsaju/togomak/v1/pkg/graph"
	"github.com/srevinsaju/togomak/v1/pkg/meta"
	"github.com/srevinsaju/togomak/v1/pkg/pipeline"
	"github.com/srevinsaju/togomak/v1/pkg/x"
	"strings"

	"github.com/zclconf/go-cty/cty"
	"os"
	"path/filepath"
	"sync"
	"time"
)

func loadGlobalParams(t *Togomak, cfg Config) {
	paramsGo := make(map[string]cty.Value)
	if cfg.Child {
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

func Orchestra(cfg Config) int {
	var diags hcl.Diagnostics
	t, ctx := NewContextWithTogomak(cfg)
	ctx, cancel := context.WithCancel(ctx)
	logger := t.Logger

	defer cancel()
	defer diagnostics(&t, &diags)

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
	if len(pipe.Imports) != 0 {
		logger.Debugf("expanding imports")
		d = pipe.Imports.PopulateProperties()
		diags = diags.Extend(d)
		if d.HasErrors() {
			return fatal(ctx)
		}
		logger.Debugf("populating properties for imports completed with %d error(s)", len(d.Errs()))
		pipe, d = pipeline.ExpandImports(ctx, pipe, t.parser)
		diags = diags.Extend(d)
		logger.Debugf("expanding imports completed with %d error(s)", len(d.Errs()))
		if d.HasErrors() {
			return fatal(ctx)
		}
	}

	/// we will first expand all local blocks
	logger.Debugf("expanding local blocks")
	locals, d := pipe.Locals.Expand()
	diags = diags.Extend(d)
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
	var depGraph *depgraph.Graph
	depGraph, d = graph.TopoSort(ctx, pipe)
	diags = diags.Extend(d)
	if diags.HasErrors() {
		return fatal(ctx)
	}

	// --> run the pipeline
	// we will now run the pipeline

	logger.Debugf("starting watchdogs and signal handlers")
	// region: interrupt handler

	handler := NewHandler(ctx)
	go handler.Interrupt()
	go handler.Kill()
	go handler.Daemons()

	// endregion: interrupt handler

	var diagsMutex sync.Mutex
	var wg sync.WaitGroup

	logger.Debugf("starting runnables")
	for _, layer := range depGraph.TopoSortedLayers() {
		// we parse the TOGOMAK_ENV file at the beginning of every layer
		// this allows us to have different environments for different layers

		togomakEnvFile := filepath.Join(t.cwd, t.tempDir, meta.OutputEnvFile)
		logger.Tracef("%s will be stored and exported here: %s", meta.OutputEnvVar, togomakEnvFile)
		envFile, err := os.OpenFile(togomakEnvFile, os.O_RDONLY|os.O_CREATE, 0644)
		if err == nil {
			e, err := envparse.Parse(envFile)
			if err != nil {
				diags = diags.Append(&hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "could not parse TOGOMAK_ENV file",
					Detail:   err.Error(),
				})
				break
			}
			x.Must(envFile.Close())
			ee := make(map[string]cty.Value)
			for k, v := range e {
				ee[k] = cty.StringVal(v)
			}
			t.ectx.Variables[ci.OutputBlock] = cty.ObjectVal(ee)
		} else {
			logger.Warnf("could not open %s file, ignoring... :%s", meta.OutputEnvVar, err)
		}

		for _, runnableId := range layer {

			var runnable ci.Block
			var ok bool

			if runnableId == meta.RootStage {
				continue
			}

			if runnableId == meta.PreStage {
				if pipe.Pre == nil {
					logger.Debugf("skipping runnable pre block %s, not defined", runnableId)
					continue
				}
				runnable = &ci.Stage{Id: runnableId, CoreStage: pipe.Pre.CoreStage}
			} else if runnableId == meta.PostStage {
				if pipe.Post == nil {
					logger.Debugf("skipping runnable post block %s, not defined", runnableId)
					continue
				}
				runnable = &ci.Stage{Id: runnableId, CoreStage: pipe.Post.CoreStage}
			} else {
				runnable, d = ci.Resolve(ctx, pipe, runnableId)
				if d.HasErrors() {
					diagsMutex.Lock()
					diags = diags.Extend(d)
					diagsMutex.Unlock()
					break
				}
			}

			logger.Debugf("runnable %s is %T", runnableId, runnable)
			handler.Tracker.AppendRunnable(runnable)

			ok, d = runnable.CanRun(ctx)
			if d.HasErrors() {
				diagsMutex.Lock()
				diags = diags.Extend(d)
				diagsMutex.Unlock()
				break
			}

			// region: requested stages, whitelisting and blacklisting
			overridden := false
			if runnable.Type() == ci.StageBlock || runnable.Type() == ci.ModuleBlock {
				stageStatus, stageStatusOk := filterList.Get(runnableId)

				// when a particular stage is explicitly requested, for example
				// in the pipeline containing the following stages
				// - hello_1
				// - hello_2
				// - hello_3
				// - hello_4 (depends on hello_1)
				// if 'hello_1' is explicitly requested, we will run 'hello_4' as well
				if filterList.HasOperationType(filter.OperationRun) && !stageStatusOk {
					isDependentOfRequestedStage := false
					for _, ss := range filterList {
						if ss.Operation == filter.OperationRun {
							if depGraph.DependsOn(runnableId, ss.RunnableId()) {
								isDependentOfRequestedStage = true
								break
							}
						}
					}

					// if this stage is not dependent on the requested stage, we will skip it
					if !isDependentOfRequestedStage {
						ok = false
					}
				}

				if stageStatusOk {
					// overridden status is shown on the build pipeline if the
					// stage is explicitly whitelisted or blacklisted
					// using the ^ or + prefix
					overridden = true
					ok = ok || stageStatus.AnyOperations(filter.OperationWhitelist)
					if stageStatus.AllOperations(filter.OperationBlacklist) {
						ok = false
					}
				}
				runnable.Set(ci.StageContextChildStatuses, stageStatus.Children(runnableId).Marshall())

			}
			// endregion: requested stages, whitelisting and blacklisting

			d := runnable.Prepare(ctx, !ok, overridden)
			if d.HasErrors() {
				diags = diags.Extend(d)
				break
			}

			if !ok {
				logger.Debugf("skipping runnable %s, condition evaluated to false", runnableId)
				continue
			}

			if runnable.IsDaemon() {
				handler.Tracker.AppendDaemon(runnable)
			} else {
				wg.Add(1)
			}

			go func(runnableId string) {
				stageDiags := runnable.Run(ctx)

				// TODO: COME BACK HERE
				handler.Tracker.AppendCompleted(runnable)

				logger.Tracef("signaling runnable %s", runnableId)

				if !stageDiags.HasErrors() {
					if runnable.IsDaemon() {
						handler.Tracker.DaemonDone()
					} else {
						wg.Done()
					}
					return
				}
				if !runnable.CanRetry() {
					logger.Debug("runnable cannot be retried")
				} else {
					logger.Infof("retrying runnable %s", runnableId)
					retryCount := 0
					retryMinBackOff := time.Duration(runnable.MinRetryBackoff()) * time.Second
					retryMaxBackOff := time.Duration(runnable.MaxRetryBackoff()) * time.Second
					retrySuccess := false
					for retryCount < runnable.MaxRetries() {
						retryCount++
						sleepDuration := time.Duration(1) * time.Second
						if runnable.RetryExponentialBackoff() {

							if retryMinBackOff*time.Duration(retryCount) > retryMaxBackOff && retryMaxBackOff > 0 {
								sleepDuration = retryMaxBackOff
							} else {
								sleepDuration = retryMinBackOff * time.Duration(retryCount)
							}
						} else {
							sleepDuration = retryMinBackOff
						}
						logger.Warnf("runnable %s failed, retrying in %s", runnableId, sleepDuration)
						time.Sleep(sleepDuration)
						sDiags := runnable.Run(ctx)
						stageDiags = append(stageDiags, sDiags...)

						if !sDiags.HasErrors() {
							retrySuccess = true
							break
						}
					}

					if !retrySuccess {
						logger.Warnf("runnable %s failed after %d retries", runnableId, retryCount)
					}

				}
				diagsMutex.Lock()
				diags = diags.Extend(stageDiags)
				diagsMutex.Unlock()
				if runnable.IsDaemon() {
					handler.Tracker.DaemonDone()
				} else {
					wg.Done()
				}

			}(runnableId)

			if cfg.Pipeline.DryRun {
				// TODO: implement --concurrency option
				// wait for the runnable to finish
				// disable concurrency
				wg.Wait()
				handler.Tracker.DaemonWait()
			}
		}
		wg.Wait()

		if diags.HasErrors() {
			if handler.Tracker.HasDaemons() && !cfg.Pipeline.DryRun && !cfg.Unattended {
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
	if diags.HasErrors() {
		return fatal(ctx)
	}
	return ok(ctx)
}
