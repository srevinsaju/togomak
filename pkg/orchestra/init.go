package orchestra

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/srevinsaju/togomak/v1/pkg/meta"
	"github.com/srevinsaju/togomak/v1/pkg/ui"
	"github.com/srevinsaju/togomak/v1/pkg/x"
	"github.com/zclconf/go-cty/cty"
	"os"
	"path/filepath"
)

func InitPipeline(dir string) {
	f := hclwrite.NewEmptyFile()
	rootBody := f.Body()
	togomakBlock := rootBody.AppendNewBlock("togomak", nil)
	togomakBlock.Body().SetAttributeValue("version", cty.NumberIntVal(1))

	// add the data block
	stageBlock := rootBody.AppendNewBlock("stage", []string{"example"})
	stageBlock.Body().SetAttributeValue("name", cty.StringVal("example"))
	stageBlock.Body().SetAttributeValue("script", cty.StringVal("echo hello world"))

	path := filepath.Join(dir, meta.ConfigFileName)
	if x.FileExists(path) {
		allow := ui.PromptYesNo("A togomak pipeline already exists in this directory. Do you want to overwrite it?")
		if !allow {
			os.Exit(1)
		}
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		ui.Error("failed to create pipeline file")
		os.Exit(1)
	}
	defer file.Close()

	n, err := f.WriteTo(file)
	if err != nil {
		ui.Error("failed to write pipeline file")
		os.Exit(1)
	}
	ui.Success("successfully wrote %d bytes to %s", n, path)

}
