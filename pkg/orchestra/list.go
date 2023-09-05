package orchestra

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/srevinsaju/togomak/v1/pkg/c"
	"github.com/srevinsaju/togomak/v1/pkg/meta"
	"github.com/srevinsaju/togomak/v1/pkg/pipeline"
	"github.com/srevinsaju/togomak/v1/pkg/ui"
	"github.com/srevinsaju/togomak/v1/pkg/x"
	"os"
	"path/filepath"
)

func List(cfg Config) error {
	logger := NewLogger(cfg)
	parser := hclparse.NewParser()

	// TODO: move this to a function
	// TODO: reduce duplication
	pipelineId := uuid.New().String()
	tmpDir := filepath.Join(meta.BuildDirPrefix, "pipelines", "tmp")
	err := os.MkdirAll(tmpDir, 0755)
	x.Must(err)
	tmpDir, err = os.MkdirTemp(tmpDir, pipelineId)
	x.Must(err)

	// TODO: move this to a function
	ctx := context.Background()
	cwd := Chdir(cfg, logger)
	ctx = context.WithValue(ctx, c.TogomakContextOwd, cfg.Owd)
	ctx = context.WithValue(ctx, c.TogomakContextCwd, cwd)
	ctx = context.WithValue(ctx, c.TogomakContextPipelineFilePath, cfg.Pipeline.FilePath)
	ctx = context.WithValue(ctx, c.TogomakContextTempDir, tmpDir)

	dgwriter := hcl.NewDiagnosticTextWriter(os.Stdout, parser.Files(), 0, true)
	pipe, hclDiags := pipeline.Read(ctx, parser)
	if hclDiags.HasErrors() {
		logger.Fatal(dgwriter.WriteDiagnostics(hclDiags))
	}

	pipe, d := pipeline.ExpandImports(ctx, pipe)
	hclDiags = hclDiags.Extend(d)

	for _, stage := range pipe.Stages {
		fmt.Println(ui.Bold(stage.Id))
	}
	return nil

}
