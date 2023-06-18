package orchestra

import (
	"context"
	"fmt"
	"github.com/kendru/darwin/go/depgraph"
	"github.com/sirupsen/logrus"
	"github.com/srevinsaju/togomak/v1/pkg/c"
	"github.com/srevinsaju/togomak/v1/pkg/ci"
	"github.com/srevinsaju/togomak/v1/pkg/diag"
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

func Orchestra(cfg Config) {
	var diags diag.Diagnostics
	t, ctx := NewContextWithTogomak(cfg)
	ctx, cancel := context.WithCancel(ctx)
	logger := t.logger

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
		logger.Fatal(err)
	}

	/// we will first expand all local blocks
	locals, d := pipe.Locals.Expand()
	if d.HasErrors() {
		d.Fatal(logger.WriterLevel(logrus.ErrorLevel))
	}
	pipe.Local = locals

	// store the pipe in the context
	ctx = context.WithValue(ctx, c.TogomakContextPipeline, pipe)

	// --> validate the pipeline
	// TODO: validate the pipeline

	// --> generate a dependency graph
	// we will now generate a dependency graph from the pipeline
	// this will be used to generate the pipeline
	var depGraph *depgraph.Graph
	depGraph, diags = graph.TopoSort(ctx, pipe)
	if diags.HasErrors() {
		diags.Fatal(logger.WriterLevel(logrus.ErrorLevel))
	}

	// --> run the pipeline
	// we will now run the pipeline

	var diagsMutex sync.Mutex
	var wg sync.WaitGroup
	var daemonWg sync.WaitGroup
	var hasDaemons bool
	var runnables ci.Runnables

	// region: interrupt handler
	chInterrupt := make(chan os.Signal, 1)
	chKill := make(chan os.Signal, 1)
	signal.Notify(chInterrupt, os.Interrupt)
	signal.Notify(chInterrupt, syscall.SIGTERM)
	signal.Notify(chKill, os.Kill)
	go InterruptHandler(ctx, cancel, chInterrupt, &runnables)
	go KillHandler(ctx, cancel, chKill, &runnables)
	// endregion: interrupt handler

	for _, layer := range depGraph.TopoSortedLayers() {
		for _, runnableId := range layer {
			var runnable ci.Runnable
			var ok bool

			if runnableId == meta.RootStage {
				continue
			}

			runnable, diags = ci.Resolve(ctx, pipe, runnableId)
			if diags.HasErrors() {
				diags.Fatal(logger.WriterLevel(logrus.ErrorLevel))
				return
			}
			logger.Debugf("runnable %s is %T", runnableId, runnable)
			runnables = append(runnables, runnable)

			ok, diags = runnable.CanRun(ctx)
			if diags.HasErrors() {
				diags.Fatal(logger.WriterLevel(logrus.ErrorLevel))
				return
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
					ok = ok || stageStatus.Operation == ConfigPipelineStageRunWhitelistOperation
					if stageStatus.Operation == ConfigPipelineStageRunBlacklistOperation {
						ok = false
					}
				}

			}
			// endregion: requested stages, whitelisting and blacklisting

			d := runnable.Prepare(ctx, !ok, overridden)
			if d.HasErrors() {
				diags.Fatal(logger.WriterLevel(logrus.ErrorLevel))
				return
			}

			if !ok {
				logger.Debugf("skipping runnable %s, condition evaluated to false", runnableId)
				continue
			}

			if runnable.IsDaemon() {
				hasDaemons = true
				daemonWg.Add(1)
			} else {
				wg.Add(1)
			}

			go func(runnableId string) {
				stageDiags := runnable.Run(ctx)
				if !stageDiags.HasErrors() {
					wg.Done()
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
			diags.Write(logger.WriterLevel(logrus.ErrorLevel))
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

		}

	}

	daemonWg.Wait()

	if diags.HasErrors() || diags.HasWarnings() {
		diags.Fatal(logger.WriterLevel(logrus.ErrorLevel))
	}
	Finale(ctx, logrus.InfoLevel)

}

func Finale(ctx context.Context, logLevel logrus.Level) {
	logger := ctx.Value(c.TogomakContextLogger).(*logrus.Logger)
	bootTime := ctx.Value(c.TogomakContextBootTime).(time.Time)
	logger.Log(logLevel, ui.Grey(fmt.Sprintf("took %s", time.Since(bootTime).Round(time.Millisecond))))
}
