package orchestra

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/srevinsaju/togomak/v1/pkg/ci"
	"github.com/srevinsaju/togomak/v1/pkg/conductor"
	"github.com/srevinsaju/togomak/v1/pkg/ui"
	"os"
)

func List(cfg conductor.Config) error {

	togomak := conductor.NewTogomak(cfg)
	logger := togomak.Logger
	ctx := togomak.Context

	dgwriter := hcl.NewDiagnosticTextWriter(os.Stdout, togomak.Parser.Files(), 0, true)
	pipe, hclDiags := ci.Read(cfg.Paths, togomak.Parser)
	if hclDiags.HasErrors() {
		logger.Fatal(dgwriter.WriteDiagnostics(hclDiags))
	}

	pipe, d := pipe.ExpandImports(ctx, togomak.Parser, togomak.Config.Paths.Cwd)
	hclDiags = hclDiags.Extend(d)

	for _, stage := range pipe.Stages {
		fmt.Println(ui.Bold(stage.Id))
	}
	return nil

}
