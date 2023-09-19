package pipeline

import (
	"context"
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/srevinsaju/togomak/v1/pkg/c"
	"github.com/srevinsaju/togomak/v1/pkg/ci"
	"github.com/srevinsaju/togomak/v1/pkg/global"
	"github.com/srevinsaju/togomak/v1/pkg/meta"
	"github.com/srevinsaju/togomak/v1/pkg/ui"
	"os"
	"path/filepath"
	"strings"
)

// ConfigFilePath returns the path to the configuration file. If the path is not absolute, it is assumed to be
// relative to the working directory
// DEPRECATED: use configFileDir instead
func ConfigFilePath(ctx context.Context) string {
	filePath := ctx.Value(c.TogomakContextPipelineFilePath).(string)
	if filePath == "" {
		filePath = meta.ConfigFileName
	}
	owd := ctx.Value(c.TogomakContextOwd).(string)

	if filepath.IsAbs(filePath) == false {
		filePath = filepath.Join(owd, filePath)
	}
	return filePath
}

func ConfigFileDir(ctx context.Context) string {
	return filepath.Dir(ConfigFilePath(ctx))
}

// Read reads togomak.hcl from the configuration file directory. A configuration file directory is the one that
// contains togomak.hcl, it searches recursively outwards.
// DEPRECATED: use ReadDir instead
func Read(ctx context.Context, parser *hclparse.Parser) (*ci.Pipeline, hcl.Diagnostics) {
	ciFile := ConfigFilePath(ctx)

	f, diags := parser.ParseHCLFile(ciFile)
	if diags.HasErrors() {
		return nil, diags
	}

	pipeline := &ci.Pipeline{}
	diags = gohcl.DecodeBody(f.Body, nil, pipeline)

	if pipeline.Builder.Version != 1 {
		return ReadDir(ctx, parser)
	} else if pipeline.Builder.Version == 1 {
		ui.DeprecationWarning(fmt.Sprintf("%s configuration version 1 is deprecated, and support for the same will be removed in a later version. ", meta.AppName))
	}
	return pipeline, diags
}

// ReadDir parses an entire directory of *.hcl files and merges them together. This is useful when you want to
// split your pipeline into multiple files, without having to import them individually
func ReadDir(ctx context.Context, parser *hclparse.Parser) (*ci.Pipeline, hcl.Diagnostics) {
	dir := ConfigFileDir(ctx)
	return ReadDirFromPath(dir, parser)

}

func ReadDirFromPath(dir string, parser *hclparse.Parser) (*ci.Pipeline, hcl.Diagnostics) {
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

		p := &ci.Pipeline{}

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

// Meta is a helper struct to create a pipeline from multiple pipelines
// this additionally includes the file pointer f, and the filename
type Meta struct {
	pipe     *ci.Pipeline
	f        *hcl.File
	filename string
}

func NewMeta(pipe *ci.Pipeline, f *hcl.File, filename string) *Meta {
	return &Meta{
		pipe:     pipe,
		f:        f,
		filename: filename,
	}
}

type MetaList []*Meta

func (m MetaList) Append(p *Meta) MetaList {
	return append(m, p)
}

func (m MetaList) Extend(p MetaList) MetaList {
	return append(m, p...)
}

// Merge creates a pipeline from multiple pipelines. This is useful when you want to merge multiple
// pipelines together, without having to import them individually
func Merge(pipelines MetaList) (*ci.Pipeline, hcl.Diagnostics) {
	pipe := &ci.Pipeline{}

	var diags hcl.Diagnostics

	var versionDefinedFromFilename string
	var pre *ci.PreStage
	var post *ci.PostStage

	for _, p := range pipelines {
		if p.pipe == nil {
			global.Logger().Debugf("pipeline %s is nil", p.filename)
			panic("pipeline is nil")
		}
		if pipe.Builder.Version == 0 && p.pipe.Builder.Version != 0 {
			pipe.Builder.Version = p.pipe.Builder.Version
			versionDefinedFromFilename = p.filename
		}
		if p.pipe.Builder.Version != pipe.Builder.Version && p.pipe.Builder.Version != 0 {
			// when overriding and using multiple pipelines, the version of the togomak pipeline schema is
			// required to be the same
			return nil, diags.Append(&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "version mismatch",
				Detail:   fmt.Sprintf("version mismatch between pipelines: %d (%s) and %d (%s)", p.pipe.Builder.Version, p.filename, pipe.Builder.Version, versionDefinedFromFilename),
			})
		}

		// TODO: create an error if there are duplicate resource definition
		if pipe.Stages.CheckIfDistinct(p.pipe.Stages).HasErrors() {
			return nil, diags.Append(&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "duplicate stage",
				Detail:   fmt.Sprintf("duplicate stage definition in %s", p.filename),
			})
		}
		if pipe.Data.CheckIfDistinct(p.pipe.Data).HasErrors() {
			return nil, diags.Append(&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "duplicate data block",
				Detail:   fmt.Sprintf("duplicate data block definition in %s", p.filename),
			})
		}
		if pipe.Local.CheckIfDistinct(p.pipe.Local).HasErrors() {
			return nil, diags.Append(&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "duplicate local",
				Detail:   fmt.Sprintf("duplicate local definition in %s", p.filename),
			})
		}
		if pipe.Macros.CheckIfDistinct(p.pipe.Macros).HasErrors() {
			return nil, diags.Append(&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "duplicate macro",
				Detail:   fmt.Sprintf("duplicate macro definition in %s", p.filename),
			})
		}

		if pipe.Modules.CheckIfDistinct(p.pipe.Modules).HasErrors() {
			return nil, diags.Append(&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "duplicate module",
				Detail:   fmt.Sprintf("duplicate module definition in %s", p.filename),
			})
		}

		if p.pipe.Pre != nil {
			if pre != nil {
				return nil, diags.Append(&hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "duplicate pre stage",
					Detail:   fmt.Sprintf("duplicate pre stage definition in %s", p.filename),
				})
			}
			pre = p.pipe.Pre
		}

		if p.pipe.Post != nil {
			if post != nil {
				return nil, diags.Append(&hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "duplicate post stage",
					Detail:   fmt.Sprintf("duplicate post stage definition in %s", p.filename),
				})
			}
			post = p.pipe.Post
		}

		pipe.Stages = append(pipe.Stages, p.pipe.Stages...)
		pipe.Data = append(pipe.Data, p.pipe.Data...)
		pipe.DataProviders = append(pipe.DataProviders, p.pipe.DataProviders...)
		pipe.Macros = append(pipe.Macros, p.pipe.Macros...)
		pipe.Modules = append(pipe.Modules, p.pipe.Modules...)
		pipe.Local = append(pipe.Local, p.pipe.Local...)
		pipe.Locals = append(pipe.Locals, p.pipe.Locals...)
		pipe.Imports = append(pipe.Imports, p.pipe.Imports...)

	}
	pipe.Pre = pre
	pipe.Post = post
	return pipe, diags
}
