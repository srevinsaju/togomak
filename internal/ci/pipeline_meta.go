package ci

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/srevinsaju/togomak/v1/internal/global"
)

// Meta is a helper struct to create a pipeline from multiple pipelines
// this additionally includes the file pointer f, and the filename
type Meta struct {
	pipe     *Pipeline
	f        *hcl.File
	filename string
}

func NewMeta(pipe *Pipeline, f *hcl.File, filename string) *Meta {
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
func Merge(pipelines MetaList) (*Pipeline, hcl.Diagnostics) {
	pipe := &Pipeline{}

	var diags hcl.Diagnostics

	var versionDefinedFromFilename string
	var pre *PreStage
	var post *PostStage

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
		if pipe.Vars.CheckIfDistinct(p.pipe.Vars).HasErrors() {
			return nil, diags.Append(&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "duplicate variable",
				Detail:   fmt.Sprintf("duplicate variable definition in %s", p.filename),
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
		pipe.Vars = append(pipe.Vars, p.pipe.Vars...)

	}
	pipe.Pre = pre
	pipe.Post = post
	return pipe, diags
}
