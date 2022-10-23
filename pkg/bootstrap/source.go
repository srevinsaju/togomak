package bootstrap

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/spf13/afero"
	"github.com/srevinsaju/togomak/pkg/context"
	"github.com/srevinsaju/togomak/pkg/meta"
	"github.com/srevinsaju/togomak/pkg/schema"
	"github.com/srevinsaju/togomak/pkg/x"

	"github.com/otiai10/copy"

	"net/url"
	"os"
	"path/filepath"
	"strconv"
)

const (
	// SourceTypeGit specifies the git source type
	SourceTypeGit = "git"

	// SourceTypeFile specifies the file source type
	SourceTypeFile = "file"
	// TODO: implement http, https, etc.
)

const (
	// SourceTypeGitConfigCloneDepth allows users to configure the clone depth
	SourceTypeGitConfigCloneDepth = "clone_depth"

	// SourceTypeGitConfigSingleBranch allows users to configure if only a single branch is cloned
	SourceTypeGitConfigSingleBranch = "single_branch"

	// SourceTypeGitConfigInsecureSkipTLS allows users to configure if TLS is skipped
	SourceTypeGitConfigInsecureSkipTLS = "insecure_skip_tls"

	// SourceTypeGitConfigDefaultsCloneDepth gives the default clone depth
	SourceTypeGitConfigDefaultsCloneDepth = "1"

	SourceTypeGitConfigReferenceName = "reference_name"
)

func ExpandSources(ctx *context.Context, data *schema.SchemaConfig) {
	// TODO: complete this
	// if the stage extends on a git repository, clone it

	for _, v := range data.Stages {
		if v.Source.Type == "" {
			continue
		}

		u, err := url.Parse(v.Source.URL)
		if err != nil {
			ctx.Logger.Fatal("Failed to parse extends parameter", err)
		}

		childCtx := ctx.AddChild(v.Source.Type, "")
		childCtx.Logger.Trace("Detected source type")

		childCtx.Logger.Trace("Creating directories")
		var dest string
		switch v.Source.Type {
		case SourceTypeGit:
			dest = filepath.Join(ctx.Data.GetString("cwd"), meta.BuildDirPrefix, meta.BuildDir, meta.ExtendsDir, SourceTypeGit, u.Host, u.Path)
		case SourceTypeFile:
			dest = filepath.Join(ctx.Data.GetString(context.KeyCwd), meta.BuildDirPrefix, meta.BuildDir, meta.ExtendsDir, SourceTypeFile, u.Path)
		}
		if exists, err := afero.Exists(afero.OsFs{}, dest); exists || err != nil {
			if err != nil {
				childCtx.Logger.Warnf("Failed to check if directory %s exists: %s", dest, err)
				return
			}
			childCtx.Logger.Infof("%s already exists, skipping.", dest)
		}
		x.Must(os.MkdirAll(dest, 0755))

		switch v.Source.Type {
		case SourceTypeGit:
			// parse clone depth, and other git clone parameters
			gitCloneDepth := childCtx.GetenvWithDefault(SourceTypeGitConfigCloneDepth, SourceTypeGitConfigDefaultsCloneDepth)
			gitCloneDepthParsed, err := strconv.Atoi(gitCloneDepth)
			if err != nil {
				childCtx.Logger.Fatalf("Invalid value provided for %s=%s: %s", SourceTypeGitConfigCloneDepth, gitCloneDepth, err)
			}
			childCtx.Logger.Tracef("Parsed data for %s=%s: %d", SourceTypeGitConfigCloneDepth, gitCloneDepth, gitCloneDepthParsed)

			isGitCloneSingleBranch := childCtx.Getenv(SourceTypeGitConfigSingleBranch) != ""
			childCtx.Logger.Tracef("Parsed data for %s: %v", SourceTypeGitConfigSingleBranch, isGitCloneSingleBranch)

			isGitCloneInsecureSkipTLS := childCtx.Getenv(SourceTypeGitConfigInsecureSkipTLS) != ""
			childCtx.Logger.Tracef("Parsed data for %s: %v", SourceTypeGitConfigInsecureSkipTLS, isGitCloneInsecureSkipTLS)

			gitCloneReferenceName := childCtx.Getenv(SourceTypeGitConfigReferenceName)
			if gitCloneReferenceName == "" {
				childCtx.Logger.Debug("Using refs/heads/master as default reference name")
				gitCloneReferenceName = "refs/heads/master"
			}

			// clone the repository
			childCtx.Logger.Debugf("Cloning %s to %s with reference %s", v.Source.URL, dest, gitCloneReferenceName)
			_, err = git.PlainClone(dest, false, &git.CloneOptions{
				URL:   v.Source.URL,
				Depth: gitCloneDepthParsed,
				// TODO: allow the user to specify the recursive depth of the clone
				RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
				SingleBranch:      isGitCloneSingleBranch,
				ReferenceName:     plumbing.ReferenceName(gitCloneReferenceName),
				InsecureSkipTLS:   isGitCloneInsecureSkipTLS,
				Progress:          childCtx.Logger.Writer(),
			})

			// FIXME: this is a hack to fix the issue of git clone not working with
			// the reference name. This is a temporary fix, and should be removed
			// once the issue is fixed upstream
			// https://github.com/go-git/go-git/issues/363
			if err != nil {
				gitCloneReferenceName = "refs/heads/main"
				childCtx.Logger.Debugf("Cloning %s to %s with reference %s", v.Source.URL, dest, gitCloneReferenceName)
				_, err = git.PlainClone(dest, false, &git.CloneOptions{
					URL:   v.Source.URL,
					Depth: gitCloneDepthParsed,
					// TODO: allow the user to specify the recursive depth of the clone
					RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
					SingleBranch:      isGitCloneSingleBranch,
					ReferenceName:     plumbing.ReferenceName(gitCloneReferenceName),
					InsecureSkipTLS:   isGitCloneInsecureSkipTLS,
					Progress:          childCtx.Logger.Writer(),
				})
				if err != nil {
					ctx.Logger.Fatal("Failed to clone repository", err)
				}
			}
		case SourceTypeFile:
			childCtx.Logger.Tracef("Copying %s to %s", v.Source.URL, dest)
			x.Must(copy.Copy(u.Path, dest))
		}
	}

}
