package bootstrap

import (
	"github.com/srevinsaju/togomak/pkg/context"
	"os"
)

func TempDir(ctx *context.Context) {
	tempDir, err := os.MkdirTemp("", ".togomak")

	if err != nil {
		ctx.Logger.Fatal(err)
	}
	ctx.TempDir = tempDir
}

func SafeDeleteTempDir(ctx *context.Context) {
	err := os.RemoveAll(ctx.TempDir)
	if err != nil {
		ctx.Logger.Warn("Failed to remove temporary directory: ", err)
	}
}
