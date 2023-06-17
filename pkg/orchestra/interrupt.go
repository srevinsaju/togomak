package orchestra

import (
	"context"
	"github.com/sirupsen/logrus"
	"github.com/srevinsaju/togomak/v1/pkg/c"
	"github.com/srevinsaju/togomak/v1/pkg/ci"
	"github.com/srevinsaju/togomak/v1/pkg/diag"
	"os"
	"os/signal"
	"time"
)

func InterruptHandler(ctx context.Context, cancel context.CancelFunc, ch chan os.Signal, runnables *ci.Runnables) {
	logger := ctx.Value(c.TogomakContextLogger).(*logrus.Logger)
	select {
	case <-ch:
		var diags diag.Diagnostics
		logger.Warn("received interrupt signal, cancelling the pipeline")
		logger.Warn("stopping running operations...")
		logger.Warn("press CTRL+C again to force quit")

		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt)
		go func() {
			<-ch
			logger.Warn("Two interrupts received. Exiting immediately.")
			diags = diags.Append(diag.Diagnostic{
				Severity: diag.SeverityError,
				Summary:  "Force quit",
				Detail:   "data loss may have occurred",
				Source:   "orchestra",
			})
			diags.Fatal(logger.WriterLevel(logrus.ErrorLevel))
			Finale(ctx, logrus.FatalLevel)
			os.Exit(1)
			return
		}()
		for _, runnable := range *runnables {
			logger.Debugf("stopping runnable %s", runnable.Identifier())
			d := runnable.Terminate()
			diags = diags.Extend(d)
		}

		if diags.HasErrors() || diags.HasWarnings() {
			diags.Write(logger.WriterLevel(logrus.ErrorLevel))
		}
		cancel()
	case <-ctx.Done():
		logger.Infof("took %s to complete the pipeline", time.Since(ctx.Value(c.TogomakContextBootTime).(time.Time)))
		return
	}
}

func KillHandler(ctx context.Context, cancel context.CancelFunc, ch chan os.Signal, runnables *ci.Runnables) {
	logger := ctx.Value(c.TogomakContextLogger).(*logrus.Logger)
	select {
	case <-ch:
		var diags diag.Diagnostics
		logger.Warn("received kill signal, killing all subprocesses")
		logger.Warn("stopping running operations...")

		for _, runnable := range *runnables {
			logger.Debugf("killing runnable %s", runnable.Identifier())
			d := runnable.Kill()
			diags = diags.Extend(d)
		}

		cancel()
		diags = diags.Append(diag.Diagnostic{
			Severity: diag.SeverityError,
			Summary:  "Force quit",
			Detail:   "data loss may have occurred",
			Source:   "orchestra",
		})
		diags.Fatal(logger.WriterLevel(logrus.ErrorLevel))
		Finale(ctx, logrus.FatalLevel)
		os.Exit(1)
	case <-ctx.Done():
		logger.Infof("took %s to complete the pipeline", time.Since(ctx.Value(c.TogomakContextBootTime).(time.Time)))
		return
	}
}
