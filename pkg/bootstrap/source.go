package bootstrap

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/spf13/afero"
	"github.com/srevinsaju/togomak/pkg/context"
	"github.com/srevinsaju/togomak/pkg/schema"
	"github.com/srevinsaju/togomak/pkg/sources"
	"github.com/srevinsaju/togomak/pkg/x"
	"strings"

	"github.com/otiai10/copy"

	"os"
	"strconv"
)

func ExpandSources(ctx *context.Context, data *schema.SchemaConfig) {
	// TODO: complete this
	// if the stage extends on a git repository, clone it

	for _, v := range data.Stages {
		if v.Source.Type == "" {
			continue
		}

		childCtx := ctx.AddChild(v.Source.Type, "")
		childCtx.Logger.Trace("Detected source type")

		childCtx.Logger.Trace("Creating directories")

		dest := sources.GetStorePath(ctx, v)

		if exists, err := afero.Exists(afero.OsFs{}, dest); exists || err != nil {
			if err != nil {
				childCtx.Logger.Warnf("Failed to check if directory %s exists: %s", dest, err)
				return
			}
			childCtx.Logger.Infof("%s already exists, skipping.", dest)
			return
		}
		x.Must(os.MkdirAll(dest, 0755))

		switch v.Source.Type {
		case sources.TypeGit:
			// parse clone depth, and other git clone parameters
			gitCloneDepth := childCtx.GetenvWithDefault(sources.TypeGitConfigCloneDepth, sources.TypeGitConfigDefaultsCloneDepth)
			gitCloneDepthParsed, err := strconv.Atoi(gitCloneDepth)
			if err != nil {
				childCtx.Logger.Fatalf("Invalid value provided for %s=%s: %s", sources.TypeGitConfigCloneDepth, gitCloneDepth, err)
			}
			childCtx.Logger.Tracef("Parsed data for %s=%s: %d", sources.TypeGitConfigCloneDepth, gitCloneDepth, gitCloneDepthParsed)

			isGitCloneSingleBranch := childCtx.Getenv(sources.TypeGitConfigSingleBranch) != ""
			childCtx.Logger.Tracef("Parsed data for %s: %v", sources.TypeGitConfigSingleBranch, isGitCloneSingleBranch)

			isGitCloneInsecureSkipTLS := childCtx.Getenv(sources.TypeGitConfigInsecureSkipTLS) != ""
			childCtx.Logger.Tracef("Parsed data for %s: %v", sources.TypeGitConfigInsecureSkipTLS, isGitCloneInsecureSkipTLS)

			gitCloneReferenceName := childCtx.Getenv(sources.TypeGitConfigReferenceName)
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
		case sources.TypeFile:
			childCtx.Logger.Tracef("Copying %s to %s", v.Source.URL, dest)
			x.Must(copy.Copy(strings.TrimPrefix("file://", v.Source.URL), dest))
		}
	}

}
