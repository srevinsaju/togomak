package orchestra

import (
	"context"
	"github.com/hashicorp/hcl/v2"
	"github.com/sirupsen/logrus"
	"github.com/srevinsaju/togomak/v1/pkg/c"
	"github.com/srevinsaju/togomak/v1/pkg/ci"
	"os"
	"os/signal"
	"time"
)

func InterruptHandler(ctx context.Context, cancel context.CancelFunc, ch chan os.Signal, runnables *ci.Blocks) {
	logger := ctx.Value(c.TogomakContextLogger).(*logrus.Logger)
	select {
	case <-ch:
		var diags hcl.Diagnostics
		logger.Warn("received interrupt signal, cancelling the pipeline")
		logger.Warn("stopping running operations...")
		logger.Warn("press CTRL+C again to force quit")

		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt)
		go func() {
			<-ch
			logger.Warn("Two interrupts received. Exiting immediately.")
			diags = diags.Append(&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Force quit",
				Detail:   "data loss may have occurred",
			})
			os.Exit(fatal(ctx))
			return
		}()
		for _, runnable := range *runnables {
			logger.Debugf("stopping runnable %s", runnable.Identifier())
			d := runnable.Terminate(false)
			diags = diags.Extend(d)
		}

		if diags.HasErrors() {
			writer := hcl.NewDiagnosticTextWriter(os.Stderr, nil, 78, true)
			_ = writer.WriteDiagnostics(diags)
			os.Exit(fatal(ctx))
		}
		cancel()
	case <-ctx.Done():
		logger.Infof("took %s to complete the pipeline", time.Since(ctx.Value(c.TogomakContextBootTime).(time.Time)))
		return
	}
}

func KillHandler(ctx context.Context, cancel context.CancelFunc, ch chan os.Signal, runnables *ci.Blocks) {
	logger := ctx.Value(c.TogomakContextLogger).(*logrus.Logger)
	select {
	case <-ch:
		var diags hcl.Diagnostics
		logger.Warn("received kill signal, killing all subprocesses")
		logger.Warn("stopping running operations...")

		for _, runnable := range *runnables {
			logger.Debugf("killing runnable %s", runnable.Identifier())
			d := runnable.Kill()
			diags = diags.Extend(d)
		}

		cancel()
		diags = diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Force quit",
			Detail:   "data loss may have occurred",
		})
		if diags.HasErrors() {
			writer := hcl.NewDiagnosticTextWriter(os.Stderr, nil, 78, true)
			_ = writer.WriteDiagnostics(diags)
		}
		os.Exit(fatal(ctx))
	case <-ctx.Done():
		logger.Infof("took %s to complete the pipeline", time.Since(ctx.Value(c.TogomakContextBootTime).(time.Time)))
		return
	}
}
