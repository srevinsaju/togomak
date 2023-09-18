package orchestra

import (
	"context"
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/sirupsen/logrus"
	"github.com/srevinsaju/togomak/v1/pkg/c"
	"github.com/srevinsaju/togomak/v1/pkg/ci"
	"github.com/srevinsaju/togomak/v1/pkg/global"
	"github.com/srevinsaju/togomak/v1/pkg/ui"
	"github.com/srevinsaju/togomak/v1/pkg/x"
	"time"
)

func Finale(ctx context.Context, logLevel logrus.Level) {
	logger := global.Logger()
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
	if diags == nil {
		return
	}
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
