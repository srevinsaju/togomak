package orchestra

import (
	"context"
	"github.com/hashicorp/hcl/v2"
	"github.com/kendru/darwin/go/depgraph"
	"github.com/sirupsen/logrus"
	"github.com/srevinsaju/togomak/v1/pkg/ci"
	"github.com/srevinsaju/togomak/v1/pkg/global"
	"github.com/srevinsaju/togomak/v1/pkg/handler"
	"github.com/srevinsaju/togomak/v1/pkg/rules"
	"github.com/srevinsaju/togomak/v1/pkg/runnable"
	"time"
)

func RunWithRetries(runnableId string, runnable ci.Block, ctx context.Context, handler *handler.Handler, logger *logrus.Logger, opts ...runnable.Option) {
	stageDiags := runnable.Run(ctx, opts...)

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

func CanRun(runnable ci.Block, ctx context.Context, filterList rules.Operations, filterQuery rules.QueryEngines, runnableId string, depGraph *depgraph.Graph, opts ...runnable.Option) (ok bool, overridden bool, diags hcl.Diagnostics) {

	ok, d := runnable.CanRun(ctx, opts...)
	if d.HasErrors() {
		diags = diags.Extend(d)
		return false, false, diags
	}

	if runnable.Type() == ci.StageBlock && len(filterQuery) != 0 {
		ok, overridden, d = filterQuery.Eval(ok, *runnable.(*ci.Stage))
		if d.HasErrors() {
			diags = diags.Extend(d)
			return false, false, diags
		}
	}
	if len(filterList) == 0 {
		filterList = append(filterList, rules.NewOperation(rules.OperationTypeAnd, "default"))
	}

	for _, rule := range filterList {
		if rule.RunnableId() == "all" {
			return ok, false, diags
		}
	}

	for _, rule := range filterList {
		if rule.RunnableId() == runnableId && rule.Operation() == rules.OperationTypeAdd {
			return true, true, diags
		}
		if rule.RunnableId() == runnableId && rule.Operation() == rules.OperationTypeSub {
			return false, true, diags
		}
		if rule.RunnableId() == runnableId && rule.Operation() == rules.OperationTypeAnd {
			return ok, false, diags
		}
		if depGraph.DependsOn(rule.RunnableId(), runnableId) {
			return true, false, diags
		}
		if runnable.Type() == ci.StageBlock {
			stage := runnable.(*ci.Stage)
			if stage.Lifecycle != nil {
				ectx := global.HclEvalContext()
				global.EvalContextMutex.RLock()
				phaseHcl, d := stage.Lifecycle.Phase.Value(ectx)
				global.EvalContextMutex.RUnlock()

				if d.HasErrors() {
					diags = diags.Extend(d)
					return false, false, diags
				}
				phases := phaseHcl.AsValueSlice()

				for _, phase := range phases {
					if rule.RunnableId() == phase.AsString() {
						return ok, false, diags
					}
				}
				if len(phases) == 0 && rule.RunnableId() == "default" {
					return ok, false, diags
				}
			} else {
				if rule.RunnableId() == "default" {
					return ok, false, diags
				}
			}
		}
	}
	if runnable.Type() == ci.StageBlock {
		return false, overridden, diags

	}
	return true, overridden, diags
}
