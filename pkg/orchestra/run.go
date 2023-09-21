package orchestra

import (
	"context"
	"github.com/hashicorp/hcl/v2"
	"github.com/kendru/darwin/go/depgraph"
	"github.com/sirupsen/logrus"
	"github.com/srevinsaju/togomak/v1/pkg/ci"
	"github.com/srevinsaju/togomak/v1/pkg/filter"
	"github.com/srevinsaju/togomak/v1/pkg/handler"
	"github.com/srevinsaju/togomak/v1/pkg/runnable"
	"time"
)

func RunWithRetries(runnableId string, runnable ci.Block, ctx context.Context, handler *handler.Handler, logger *logrus.Logger, opts ...runnable.Option) {
	stageDiags := runnable.Run(ctx)

	handler.Tracker.AppendCompleted(runnable)
	logger.Tracef("signaling runnable %s", runnableId)

	if !stageDiags.HasErrors() {
		if runnable.IsDaemon() {
			handler.Tracker.DaemonDone()
		} else {
			handler.Tracker.RunnableDone()
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
			sDiags := runnable.Run(ctx, opts...)
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
	handler.Diags.Extend(stageDiags)
	if runnable.IsDaemon() {
		handler.Tracker.DaemonDone()
	} else {
		handler.Tracker.RunnableDone()
	}
}

func CanRun(runnable ci.Block, ctx context.Context, filterList filter.FilterList, runnableId string, depGraph *depgraph.Graph, opts ...runnable.Option) (bool, hcl.Diagnostics, bool) {
	var diags hcl.Diagnostics

	ok, d := runnable.CanRun(ctx, opts...)
	if d.HasErrors() {
		diags = diags.Extend(d)
		return false, diags, false
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
	return ok, diags, overridden
}
