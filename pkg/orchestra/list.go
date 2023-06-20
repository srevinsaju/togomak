package orchestra

import (
	"fmt"
	"github.com/srevinsaju/togomak/v1/pkg/pipeline"
	"github.com/srevinsaju/togomak/v1/pkg/ui"
)

func List(cfg Config) error {
	logger := NewLogger(cfg)
	togomak, ctx := NewContextWithTogomak(cfg)

	pipe, hclDiags := pipeline.Read(ctx, togomak.parser)
	if hclDiags.HasErrors() {
		logger.Fatal(togomak.hclDiagWriter.WriteDiagnostics(hclDiags))
	}

	for _, stage := range pipe.Stages {
		fmt.Println(ui.Bold(stage.Id))
	}
	return nil

}
