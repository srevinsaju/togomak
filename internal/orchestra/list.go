package orchestra

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/srevinsaju/togomak/v1/internal/ci"
	"github.com/srevinsaju/togomak/v1/internal/ui"
	"os"
)

func List(cfg ci.ConductorConfig) error {

	conductor := ci.NewConductor(cfg)
	logger := conductor.Logger()

	dgwriter := hcl.NewDiagnosticTextWriter(os.Stdout, conductor.Parser.Files(), 0, true)
	pipe, hclDiags := ci.Read(conductor)
	if hclDiags.HasErrors() {
		logger.Fatal(dgwriter.WriteDiagnostics(hclDiags))
	}

	pipe, d := pipe.ExpandImports(conductor, conductor.Config.Paths.Cwd)
	hclDiags = hclDiags.Extend(d)

	for _, stage := range pipe.Stages {
		fmt.Println(ui.Bold(stage.Id))
	}
	return nil

}
