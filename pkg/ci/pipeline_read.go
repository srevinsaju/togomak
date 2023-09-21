package ci

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/srevinsaju/togomak/v1/pkg/global"
	"github.com/srevinsaju/togomak/v1/pkg/meta"
	"github.com/srevinsaju/togomak/v1/pkg/parse"
	"github.com/srevinsaju/togomak/v1/pkg/path"
	"github.com/srevinsaju/togomak/v1/pkg/ui"
	"os"
	"path/filepath"
	"strings"
)

// Read reads togomak.hcl from the configuration file directory. A configuration file directory is the one that
// contains togomak.hcl, it searches recursively outwards.
// DEPRECATED: use ReadDir instead
func Read(paths path.Path, parser *hclparse.Parser) (*Pipeline, hcl.Diagnostics) {
	ciFile := parse.ConfigFilePath(paths)

	f, diags := parser.ParseHCLFile(ciFile)
	if diags.HasErrors() {
		return nil, diags
	}

	pipeline := &Pipeline{}
	diags = gohcl.DecodeBody(f.Body, nil, pipeline)

	if pipeline.Builder.Version != 1 {
		return ReadDir(paths, parser)
	} else if pipeline.Builder.Version == 1 {
		ui.DeprecationWarning(fmt.Sprintf("%s configuration version 1 is deprecated, and support for the same will be removed in a later version. ", meta.AppName))
	}
	return pipeline, diags
}

// ReadDir parses an entire directory of *.hcl files and merges them together. This is useful when you want to
// split your pipeline into multiple files, without having to import them individually
func ReadDir(paths path.Path, parser *hclparse.Parser) (*Pipeline, hcl.Diagnostics) {
	dir := parse.ConfigFileDir(paths)
	return ReadDirFromPath(dir, parser)

}

func ReadDirFromPath(dir string, parser *hclparse.Parser) (*Pipeline, hcl.Diagnostics) {
	logger := global.Logger()
	var diags hcl.Diagnostics
	togomakFiles, err := os.ReadDir(dir)
	if err != nil {
		panic(err)
	}
	var pipes []*Meta
	for _, file := range togomakFiles {
		if file.IsDir() {
			continue
		}
		if !strings.HasSuffix(file.Name(), ".hcl") {
			continue
		}
		if strings.Contains(file.Name(), ".lock.hcl") {
			// we will not process .lock.hcl files
			continue
		}
		f, d := parser.ParseHCLFile(filepath.Join(dir, file.Name()))
		diags = diags.Extend(d)

		p := &Pipeline{}

		d = gohcl.DecodeBody(f.Body, nil, p)
		diags = diags.Extend(d)
		if d.HasErrors() {
			logger.Debugf("error parsing %s", file.Name())
			continue
		}
		pipes = append(pipes, &Meta{
			pipe:     p,
			f:        f,
			filename: file.Name(),
		})

	}
	pipe, d := Merge(pipes)
	diags = diags.Extend(d)
	return pipe, diags
}
