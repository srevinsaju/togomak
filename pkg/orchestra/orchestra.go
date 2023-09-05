package orchestra

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-envparse"
	"github.com/hashicorp/hcl/v2"
	"github.com/kendru/darwin/go/depgraph"
	"github.com/sirupsen/logrus"
	"github.com/srevinsaju/togomak/v1/pkg/c"
	"github.com/srevinsaju/togomak/v1/pkg/ci"
	"github.com/srevinsaju/togomak/v1/pkg/graph"
	"github.com/srevinsaju/togomak/v1/pkg/meta"
	"github.com/srevinsaju/togomak/v1/pkg/pipeline"
	"github.com/srevinsaju/togomak/v1/pkg/ui"
	"github.com/srevinsaju/togomak/v1/pkg/x"

	"github.com/zclconf/go-cty/cty"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
)

func NewLogger(cfg Config) *logrus.Logger {
	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:    false,
		DisableTimestamp: cfg.Child,
	})
	switch cfg.Verbosity {
	case -1:
	case 0:
		logger.SetLevel(logrus.InfoLevel)
		break
	case 1:
		logger.SetLevel(logrus.DebugLevel)
		break
	default:
		logger.SetLevel(logrus.TraceLevel)
		break
	}
	if cfg.Ci {
		logger.SetFormatter(&logrus.TextFormatter{
			DisableColors:             false,
			EnvironmentOverrideColors: false,
			ForceColors:               true,
			ForceQuote:                false,
		})
	}
	if cfg.Child {
		logger.SetFormatter(&logrus.TextFormatter{
			DisableTimestamp:          true,
			DisableColors:             false,
			EnvironmentOverrideColors: false,
			ForceColors:               true,
			ForceQuote:                false,
		})
	}
	return logger
}

func Chdir(cfg Config, logger *logrus.Logger) string {
	cwd := cfg.Dir
	if cwd == "" {
		cwd = filepath.Dir(cfg.Pipeline.FilePath)
		if filepath.Base(cwd) == meta.BuildDirPrefix {
			cwd = filepath.Dir(cwd)
		}
	}
	err := os.Chdir(cwd)
	if err != nil {
		logger.Fatal(err)
	}
	cwd, err = os.Getwd()
	x.Must(err)
	return cwd

}

func Orchestra(cfg Config) int {
	var diags hcl.Diagnostics
	t, ctx := NewContextWithTogomak(cfg)
	ctx, cancel := context.WithCancel(ctx)
	logger := t.Logger

	defer cancel()
	defer diagnostics(&t, &diags)

	// region: external parameters
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
	t.ectx.Variables["param"] = cty.ObjectVal(paramsGo)
	// endregion

	// --> parse the config file
	// we will now read the pipeline from togomak.hcl
	pipe, hclDiags := pipeline.Read(ctx, t.parser)
	if hclDiags.HasErrors() {
		logger.Fatal(t.hclDiagWriter.WriteDiagnostics(hclDiags))
	}

	// whitelist all stages if unspecified
	stageStatuses := cfg.Pipeline.Stages

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

	/// we will first expand all local blocks
	locals, d := pipe.Locals.Expand()
	diags = diags.Extend(d)
	if d.HasErrors() {
		return fatal(ctx)
	}

	pipe.Local = locals
	if len(pipe.Imports) != 0 {
		pipe, d = pipeline.ExpandImports(ctx, pipe)
		diags = diags.Extend(d)
		if d.HasErrors() {
			return fatal(ctx)
		}
	}

	// store the pipe in the context
	ctx = context.WithValue(ctx, c.TogomakContextPipeline, pipe)

	// --> validate the pipeline
	// TODO: validate the pipeline

	// --> generate a dependency graph
	// we will now generate a dependency graph from the pipeline
	// this will be used to generate the pipeline
	var depGraph *depgraph.Graph
	depGraph, d = graph.TopoSort(ctx, pipe)
	diags = diags.Extend(d)
	if diags.HasErrors() {
		return fatal(ctx)
	}

	// --> run the pipeline
	// we will now run the pipeline

	var diagsMutex sync.Mutex
	var wg sync.WaitGroup
	var daemonWg sync.WaitGroup
	var hasDaemons bool
	var runnables ci.Blocks
	var daemons ci.Blocks
	var daemonsMutex sync.Mutex
	var completedRunnablesSignal = make(chan ci.Block, 1)
	var completedRunnables ci.Blocks
	var completedRunnablesMutex sync.Mutex

	// region: interrupt handler
	chInterrupt := make(chan os.Signal, 1)
	chKill := make(chan os.Signal, 1)
	signal.Notify(chInterrupt, os.Interrupt)
	signal.Notify(chInterrupt, syscall.SIGTERM)
	signal.Notify(chKill, os.Kill)
	go InterruptHandler(ctx, cancel, chInterrupt, &runnables)
	go KillHandler(ctx, cancel, chKill, &runnables)
	go daemonKiller(ctx, completedRunnablesSignal, &daemons)
	// endregion: interrupt handler

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

			runnable, d = ci.Resolve(ctx, pipe, runnableId)
			if d.HasErrors() {
				diagsMutex.Lock()
				diags = diags.Extend(d)
				diagsMutex.Unlock()
				break
			}
			logger.Debugf("runnable %s is %T", runnableId, runnable)
			runnables = append(runnables, runnable)

			ok, d = runnable.CanRun(ctx)
			if d.HasErrors() {
				diagsMutex.Lock()
				diags = diags.Extend(d)
				diagsMutex.Unlock()
				break
			}

			// region: requested stages, whitelisting and blacklisting
			overridden := false
			if runnable.Type() == ci.StageBlock {
				stageStatus, stageStatusOk := stageStatuses.Get(runnableId)

				// when a particular stage is explicitly requested, for example
				// in the pipeline containing the following stages
				// - hello_1
				// - hello_2
				// - hello_3
				// - hello_4 (depends on hello_1)
				// if 'hello_1' is explicitly requested, we will run 'hello_4' as well
				if stageStatuses.HasOperationType(ConfigPipelineStageRunOperation) && !stageStatusOk {
					isDependentOfRequestedStage := false
					for _, ss := range stageStatuses {
						if ss.Operation == ConfigPipelineStageRunOperation {
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
					ok = ok || stageStatus.AnyOperations(ConfigPipelineStageRunWhitelistOperation)
					if stageStatus.AllOperations(ConfigPipelineStageRunBlacklistOperation) {
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
				hasDaemons = true
				daemonWg.Add(1)

				daemonsMutex.Lock()
				daemons = append(daemons, runnable)
				daemonsMutex.Unlock()
			} else {
				wg.Add(1)
			}

			go func(runnableId string) {
				stageDiags := runnable.Run(ctx)

				logger.Tracef("locking completedRunnablesMutex for runnable %s", runnableId)
				completedRunnablesMutex.Lock()
				completedRunnables = append(completedRunnables, runnable)
				completedRunnablesMutex.Unlock()
				logger.Tracef("unlocking completedRunnablesMutex for runnable %s", runnableId)
				completedRunnablesSignal <- runnable
				logger.Tracef("signaling runnable %s", runnableId)

				if !stageDiags.HasErrors() {
					if runnable.IsDaemon() {
						daemonWg.Done()
					} else {
						wg.Done()
					}
					return
				}
				if !runnable.CanRetry() {
					logger.Error(stageDiags.Error())
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
					daemonWg.Done()
				} else {
					wg.Done()
				}

			}(runnableId)

			if cfg.Pipeline.DryRun {
				// TODO: implement --concurrency option
				// wait for the runnable to finish
				// disable concurrency
				wg.Wait()
				daemonWg.Wait()
			}
		}
		wg.Wait()

		if diags.HasErrors() {
			if hasDaemons && !cfg.Pipeline.DryRun && !cfg.Unattended {
				logger.Info("pipeline failed, waiting for daemons to shut down")
				logger.Info("hit Ctrl+C to force stop them")
				// wait for daemons to stop
				daemonWg.Wait()
			} else if hasDaemons && !cfg.Pipeline.DryRun {
				logger.Info("pipeline failed, waiting for daemons to shut down...")
				// wait for daemons to stop
				cancel()
			}
			break
		}
	}

	daemonWg.Wait()
	if diags.HasErrors() {
		return fatal(ctx)
	}
	return ok(ctx)
}

func Finale(ctx context.Context, logLevel logrus.Level) {
	logger := ctx.Value(c.TogomakContextLogger).(*logrus.Logger)
	bootTime := ctx.Value(c.TogomakContextBootTime).(time.Time)
	logger.Log(logLevel, ui.Grey(fmt.Sprintf("took %s", time.Since(bootTime).Round(time.Millisecond))))
}

func fatal(ctx context.Context) int {
	Finale(ctx, logrus.ErrorLevel)
	return 1
}

func ok(ctx context.Context) int {
	Finale(ctx, logrus.InfoLevel)
	return 0
}

func diagnostics(t *Togomak, diags *hcl.Diagnostics) {
	x.Must(t.hclDiagWriter.WriteDiagnostics(*diags))
}

func daemonKiller(ctx context.Context, completed chan ci.Block, daemons *ci.Blocks) {
	logger := ctx.Value(c.TogomakContextLogger).(*logrus.Logger).WithField("watchdog", "")
	var completedRunnables ci.Blocks
	var diags hcl.Diagnostics
	defer diagnostics(ctx.Value(c.Togomak).(*Togomak), &diags)
	logger.Tracef("starting watchdog")

	// execute the following function when we receive any message on the completed channel
	for {
		c := <-completed
		logger.Debugf("received completed runnable, %s", c.Identifier())
		completedRunnables = append(completedRunnables, c)

		daemons := *daemons
		for _, daemon := range daemons {
			if daemon.Terminated() {
				continue
			}
			logger.Tracef("checking daemon %s", daemon.Identifier())
			lifecycle, d := daemon.Lifecycle(ctx)
			if d.HasErrors() {
				diags = diags.Extend(d)
				d := daemon.Terminate(false)
				diags = diags.Extend(d)
				return
			}
			if lifecycle == nil {
				continue
			}

			allCompleted := true
			for _, block := range lifecycle.StopWhenComplete {
				logger.Tracef("checking daemon %s, requires block %s to complete", daemon.Identifier(), block.Identifier())
				completed := false
				for _, completedBlocks := range completedRunnables {
					if block.Identifier() == completedBlocks.Identifier() {
						completed = true
						break
					}
				}
				if !completed {
					allCompleted = false
					break
				}
			}
			if allCompleted {
				logger.Infof("stopping daemon %s", daemon.Identifier())
				d := daemon.Terminate(true)
				if d.HasErrors() {
					diags = diags.Extend(d)
				}
			}
		}
	}
}
