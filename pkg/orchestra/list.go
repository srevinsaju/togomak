package orchestra

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/srevinsaju/togomak/v1/pkg/ci"
	"github.com/srevinsaju/togomak/v1/pkg/conductor"
	"github.com/srevinsaju/togomak/v1/pkg/global"
	"github.com/srevinsaju/togomak/v1/pkg/meta"
	"github.com/srevinsaju/togomak/v1/pkg/ui"
	"github.com/srevinsaju/togomak/v1/pkg/x"
	"os"
	"path/filepath"
)

func List(cfg conductor.Config) error {
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
	global.SetTempDir(tmpDir)

	// TODO: move this to a function
	ctx := context.Background()
	cwd := Chdir(cfg, logger)

	dgwriter := hcl.NewDiagnosticTextWriter(os.Stdout, parser.Files(), 0, true)
	pipe, hclDiags := ci.Read(cfg.Paths, parser)
	if hclDiags.HasErrors() {
		logger.Fatal(dgwriter.WriteDiagnostics(hclDiags))
	}

	pipe, d := pipe.ExpandImports(ctx, parser, cwd)
	hclDiags = hclDiags.Extend(d)

	for _, stage := range pipe.Stages {
		fmt.Println(ui.Bold(stage.Id))
	}
	return nil

}
