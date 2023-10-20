package orchestra

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/srevinsaju/togomak/v1/internal/ci"
	"github.com/srevinsaju/togomak/v1/internal/ui"
	"os"
)

func List(cfg ci.Config) error {

	conductor := ci.NewConductor(cfg)
	logger := conductor.Logger()
	ctx := conductor.Context()

	dgwriter := hcl.NewDiagnosticTextWriter(os.Stdout, conductor.Parser.Files(), 0, true)
	pipe, hclDiags := ci.Read(cfg.Paths, conductor.Parser)
	if hclDiags.HasErrors() {
		logger.Fatal(dgwriter.WriteDiagnostics(hclDiags))
	}

	pipe, d := pipe.ExpandImports(ctx, conductor.Parser, conductor.Config.Paths.Cwd)
	hclDiags = hclDiags.Extend(d)

	for _, stage := range pipe.Stages {
		fmt.Println(ui.Bold(stage.Id))
	}
	return nil

}
