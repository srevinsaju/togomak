package orchestra

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/srevinsaju/togomak/v1/pkg/c"
	"github.com/srevinsaju/togomak/v1/pkg/global"
	"github.com/srevinsaju/togomak/v1/pkg/ui"
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
