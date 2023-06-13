package orchestra

import (
	"context"
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/srevinsaju/togomak/v1/pkg/c"
	"github.com/srevinsaju/togomak/v1/pkg/pipeline"
	"github.com/srevinsaju/togomak/v1/pkg/ui"
	"os"
)

func List(cfg Config) error {
	logger := NewLogger(cfg)
	parser := hclparse.NewParser()
	ctx := context.Background()
	cwd := Chdir(cfg, logger)
	ctx = context.WithValue(ctx, c.TogomakContextOwd, cfg.Owd)
	ctx = context.WithValue(ctx, c.TogomakContextCwd, cwd)
	ctx = context.WithValue(ctx, c.TogomakContextPipelineFilePath, cfg.Pipeline.FilePath)

	dgwriter := hcl.NewDiagnosticTextWriter(os.Stdout, parser.Files(), 0, true)
	pipe, hclDiags := pipeline.Read(ctx, parser)
	if hclDiags.HasErrors() {
		logger.Fatal(dgwriter.WriteDiagnostics(hclDiags))
	}

	for _, stage := range pipe.Stages {
		fmt.Println(ui.Bold(stage.Id))
	}
	return nil

}
