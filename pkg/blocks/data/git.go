package data

import (
	"bytes"
	"code.gitea.io/gitea/modules/git"
	"context"
	"errors"
	"fmt"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/go-git/go-git/v5/storage"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/hashicorp/hcl/v2"
	"github.com/jdxcode/netrc"
	"github.com/sirupsen/logrus"
	"github.com/srevinsaju/togomak/v1/pkg/c"
	"github.com/srevinsaju/togomak/v1/pkg/ui"
	"github.com/zclconf/go-cty/cty"
	"io"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

type gitProviderAuthConfig struct {
	username string
	password string

	isSsh         bool
	sshPassword   string
	sshPrivateKey string
}

type gitProviderConfig struct {
	repo        string
	tag         string
	branch      string
	destination string
	commit      string

	depth    int
	caBundle []byte

	auth  gitProviderAuthConfig
	files []string
}

const (
	GitBlockArgumentUrl         = "url"
	GitBlockArgumentTag         = "tag"
	GitBlockArgumentBranch      = "branch"
	GitBlockArgumentDestination = "destination"
	GitBlockArgumentCommit      = "commit"
	GitBlockArgumentDepth       = "depth"
	GitBlockArgumentCaBundle    = "ca_bundle"
	GitBlockArgumentAuth        = "auth"
	GitBlockArgumentFiles       = "files"

	GitBlockAttrLastTag             = "last_tag"
	GitBlockAttrCommitsSinceLastTag = "commits_since_last_tag"
	GitBlockAttrSha                 = "sha"
	GitBlockAttrRef                 = "ref"

	GitBlockAttrIsTag    = "is_tag"
	GitBlockAttrIsBranch = "is_branch"
	GitBlockAttrIsNote   = "is_note"
	GitBlockAttrIsRemote = "is_remote"
)

type GitProvider struct {
	initialized bool
	Default     hcl.Expression `hcl:"default" json:"default"`

	ctx context.Context
	cfg gitProviderConfig
}

func (e *GitProvider) Name() string {
	return "git"
}

func (e *GitProvider) Identifier() string {
	return "data.git"
}

func (e *GitProvider) SetContext(context context.Context) {
	e.ctx = context
}

func (e *GitProvider) Version() string {
	return "1"
}

func (e *GitProvider) Url() string {
	return "embedded::togomak.srev.in/providers/data/git"
}

func (e *GitProvider) DecodeBody(body hcl.Body) hcl.Diagnostics {
	if !e.initialized {
		panic("provider not initialized")
	}
	var diags hcl.Diagnostics
	hclContext := e.ctx.Value(c.TogomakContextHclEval).(*hcl.EvalContext)

	schema := e.Schema()
	content, d := body.Content(schema)
	diags = diags.Extend(d)

	repo, d := content.Attributes[GitBlockArgumentUrl].Expr.Value(hclContext)
	diags = diags.Extend(d)

	tagAttr, ok := content.Attributes[GitBlockArgumentTag]
	tag := cty.StringVal("")
	if ok {
		tag, d = tagAttr.Expr.Value(hclContext)
		diags = diags.Extend(d)
	}

	branchAttr, ok := content.Attributes[GitBlockArgumentBranch]
	branch := cty.StringVal("")
	if ok {
		branch, d = branchAttr.Expr.Value(hclContext)
		diags = diags.Extend(d)
	}

	commitAttr, ok := content.Attributes[GitBlockArgumentCommit]
	commit := cty.StringVal("")
	if ok {
		commit, d = commitAttr.Expr.Value(hclContext)
		diags = diags.Extend(d)
	}

	destinationAttr, ok := content.Attributes[GitBlockArgumentDestination]
	destination := cty.StringVal("")
	if ok {
		destination, d = destinationAttr.Expr.Value(hclContext)
		diags = diags.Extend(d)
	}

	depthAttr, ok := content.Attributes[GitBlockArgumentDepth]
	depth := cty.NumberIntVal(0)
	if ok {
		depth, d = depthAttr.Expr.Value(hclContext)
		diags = diags.Extend(d)
	}

	caBundleAttr, ok := content.Attributes[GitBlockArgumentCaBundle]
	caBundle := cty.StringVal("")
	if ok {
		caBundle, d = caBundleAttr.Expr.Value(hclContext)
		diags = diags.Extend(d)
	}

	filesAttr, ok := content.Attributes[GitBlockArgumentFiles]
	files := cty.ListValEmpty(cty.String)
	if ok {
		files, d = filesAttr.Expr.Value(hclContext)
		diags = diags.Extend(d)
	}

	authBlock := content.Blocks.OfType(GitBlockArgumentAuth)
	var authConfig gitProviderAuthConfig
	if len(authBlock) == 1 {
		auth, d := content.Blocks.OfType(GitBlockArgumentAuth)[0].Body.Content(GitProviderAuthSchema())
		diags = diags.Extend(d)

		authUsername, d := auth.Attributes["username"].Expr.Value(hclContext)
		diags = diags.Extend(d)

		authPassword, d := auth.Attributes["password"].Expr.Value(hclContext)
		diags = diags.Extend(d)

		authSshPassword, d := auth.Attributes["ssh_password"].Expr.Value(hclContext)
		diags = diags.Extend(d)

		authSshPrivateKey, d := auth.Attributes["ssh_private_key"].Expr.Value(hclContext)
		diags = diags.Extend(d)

		authConfig = gitProviderAuthConfig{
			username:      authUsername.AsString(),
			password:      authPassword.AsString(),
			sshPassword:   authSshPassword.AsString(),
			sshPrivateKey: authSshPrivateKey.AsString(),
			isSsh:         authSshPassword.AsString() != "" || authSshPrivateKey.AsString() != "",
		}
	}

	depthInt, _ := depth.AsBigFloat().Int64()
	var f []string
	for _, file := range files.AsValueSlice() {
		f = append(f, file.AsString())
	}

	e.cfg = gitProviderConfig{
		repo:        repo.AsString(),
		tag:         tag.AsString(),
		branch:      branch.AsString(),
		commit:      commit.AsString(),
		destination: destination.AsString(),
		depth:       int(depthInt),
		caBundle:    []byte(caBundle.AsString()),
		auth:        authConfig,
		files:       f,
	}

	return diags
}

func (e *GitProvider) New() Provider {
	return &GitProvider{
		initialized: true,
	}
}

func GitProviderAuthSchema() *hcl.BodySchema {
	return &hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{
			{
				Name:     "username",
				Required: false,
			},
			{
				Name:     "password",
				Required: false,
			},
			{
				Name:     "ssh_password",
				Required: false,
			},
			{
				Name:     "ssh_private_key",
				Required: false,
			},
		},
	}
}

func (e *GitProvider) Schema() *hcl.BodySchema {
	return &hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{
				Type: GitBlockArgumentAuth,
			},
		},
		Attributes: []hcl.AttributeSchema{
			{
				Name:     GitBlockArgumentUrl,
				Required: true,
			},
			{
				Name:     GitBlockArgumentTag,
				Required: false,
			},
			{
				Name:     GitBlockArgumentBranch,
				Required: false,
			},
			{
				Name:     GitBlockArgumentCommit,
				Required: false,
			},
			{
				Name:     GitBlockArgumentDestination,
				Required: false,
			},
			{
				Name:     GitBlockArgumentDepth,
				Required: false,
			},
			{
				Name:     GitBlockArgumentFiles,
				Required: false,
			},
		},
	}
}

func (e *GitProvider) Initialized() bool {
	return e.initialized
}

func (e *GitProvider) Value(ctx context.Context, id string) (string, hcl.Diagnostics) {
	if !e.initialized {
		panic("provider not initialized")
	}
	return "", nil
}

func (e *GitProvider) Attributes(ctx context.Context) (map[string]cty.Value, hcl.Diagnostics) {
	logger := ctx.Value(c.TogomakContextLogger).(*logrus.Logger).WithField("provider", e.Name())
	var diags hcl.Diagnostics
	if !e.initialized {
		panic("provider not initialized")
	}

	var attrs = make(map[string]cty.Value)

	// clone git repo
	// git clone
	var s storage.Storer
	var authMethod transport.AuthMethod
	repoUrl := e.cfg.repo

	if e.cfg.auth.password != "" {
		authMethod = &http.BasicAuth{
			Username: e.cfg.auth.username,
			Password: e.cfg.auth.password,
		}
	} else if e.cfg.auth.isSsh {
		publicKeys, err := ssh.NewPublicKeys(e.cfg.auth.username, []byte(e.cfg.auth.sshPrivateKey), e.cfg.auth.sshPassword)
		if err != nil {
			diags = diags.Append(&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "ssh key error",
				Detail:   err.Error(),
			})
			return nil, diags
		}
		authMethod = publicKeys
	} else {
		// fallback to inferring it from the environment
		u, err := url.Parse(repoUrl)
		if err != nil {
			diags = diags.Append(&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "url parse error",
				Detail:   err.Error(),
			})
			return nil, diags
		}
		if u.Scheme == "http" || u.Scheme == "https" {
			usr, err := user.Current()
			if err != nil {
				panic(err)
			}
			homeDir := usr.HomeDir
			gitCredentialsFilePath := filepath.Join(homeDir, ".git-credentials")
			gitCredentialsFilePathDir := filepath.Join(homeDir, ".git", "credentials")
			netrcFilePath := filepath.Join(homeDir, ".netrc")
			if _, err := os.Stat(gitCredentialsFilePath); err == nil {
				// TODO: we will replace this will the implicit auth method
				// once go-git implements it
				authMethod = parseGitCredentialsFile(logger, gitCredentialsFilePath, u)
			} else if _, err := os.Stat(gitCredentialsFilePathDir); err == nil {
				authMethod = parseGitCredentialsFile(logger, gitCredentialsFilePathDir, u)
			} else if _, err := os.Stat(netrcFilePath); err == nil {
				authMethod = parseNetrcFile(logger, netrcFilePath, u)
			}
		} else if u.Scheme == "ssh" || u.Scheme == "git" || u.Scheme == "" {
			authMethod = nil
		}

	}

	cloneOptions := &git.CloneOptions{
		Tags:              git.AllTags,
		Depth:             e.cfg.depth,
		CABundle:          e.cfg.caBundle,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		Auth:              authMethod,
		URL:               repoUrl,
		Progress:          nil,
	}
	var repo *git.Repository
	var err error
	var cloneComplete = make(chan bool)
	go func() {
		pb := ui.NewProgressWriter(logger, fmt.Sprintf("pulling git repo %s", e.Identifier()))
		for {
			select {
			case <-cloneComplete:
				pb.Close()
				return
			default:
				pb.Write([]byte("1"))
			}
		}
	}()

	if e.cfg.destination == "" || e.cfg.destination == "memory" {
		logger.Debug("cloning into memory storage")
		s = memory.NewStorage()
		repo, err = git.CloneContext(ctx, s, memfs.New(), cloneOptions)
	} else {
		logger.Debugf("cloning to %s", e.cfg.destination)
		repo, err = git.PlainCloneContext(ctx, e.cfg.destination, false, cloneOptions)
	}
	cloneComplete <- true

	if err != nil {
		diags = diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "git clone failed",
			Detail:   err.Error(),
		})
		return nil, diags
	}

	w, err := repo.Worktree()
	if err != nil {
		diags = diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "checkout failed",
			Detail:   err.Error(),
		})
		return nil, diags
	}

	commitIter, err := repo.Log(&git.LogOptions{
		Order: git.LogOrderCommitterTime,
	})

	count := 0
	lastTag := ""
	if err != nil {
		diags = diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "git log failed",
			Detail:   err.Error(),
		})
	} else {
		latestTag, latestTagCommit, err := GetLatestTagFromRepository(repo)
		if err != nil {
			diags = diags.Append(&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "git tags failed",
				Detail:   err.Error(),
			})
			return nil, diags
		}
		if latestTag != nil {
			lastTag = latestTag.Name().Short()
		}

		logger.Debugf("latest tag is %s", lastTag)
		logger.Debugf("iterating over commits...")
		foundTagError := errors.New("found tag")

		if latestTagCommit != nil {
			commitIterErr := commitIter.ForEach(func(commit *object.Commit) error {
				ref, err := repo.Reference(latestTag.Name(), true)
				if err != nil {
					return err
				}

				commitRef, err := repo.CommitObject(ref.Hash())
				if err != nil {
					return err
				}

				fmt.Println("checking commits", commit.Hash.String(), "==", latestTagCommit.Hash.String(), ref.Hash().String(), commitRef.Hash.String())
				if latestTagCommit.Hash == commit.Hash {
					return foundTagError
				}
				count++
				return nil
			})
			if foundTagError != commitIterErr && commitIterErr != nil {
				logger.Warn(commitIterErr)
			}
		}
	}
	commitsSinceLastTag := cty.NumberIntVal(int64(count))

	var files = make(map[string]cty.Value)
	for _, f := range e.cfg.files {
		_, err := w.Filesystem.Stat(f)
		if err != nil {
			diags = diags.Append(&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "git file search failed",
				Detail:   err.Error(),
			})
			continue
		}
		file, err := w.Filesystem.Open(f)
		if err != nil {
			diags = diags.Append(&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "git file open failed",
				Detail:   err.Error(),
			})
			continue
		}
		defer file.Close()

		data, err := io.ReadAll(file)
		if err != nil {
			diags = diags.Append(&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "git file read failed",
				Detail:   err.Error(),
			})
			continue
		}
		files[f] = cty.StringVal(string(data[:]))
	}

	if files == nil || len(files) == 0 {
		attrs[GitBlockArgumentFiles] = cty.MapValEmpty(cty.String)
	} else {
		attrs[GitBlockArgumentFiles] = cty.MapVal(files)
	}

	attrs[GitBlockArgumentUrl] = cty.StringVal(e.cfg.repo)
	attrs[GitBlockArgumentTag] = cty.StringVal(e.cfg.tag)

	ref, err := repo.Head()
	refString := cty.StringVal("")
	branch := cty.StringVal("")
	tag := cty.StringVal("")
	isBranch := cty.False
	isTag := cty.False
	isRemote := cty.False
	isNote := cty.False
	sha := cty.StringVal("")

	if err != nil {
		diags = diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "git head failed",
			Detail:   err.Error(),
		})

	} else {
		refString = cty.StringVal(ref.Name().String())
		branch = cty.StringVal(ref.Name().Short())
		isBranch = cty.BoolVal(ref.Name().IsBranch())
		isTag = cty.BoolVal(ref.Name().IsTag())
		tag = cty.StringVal(ref.Name().Short())
		isRemote = cty.BoolVal(ref.Name().IsRemote())
		isNote = cty.BoolVal(ref.Name().IsNote())
		sha = cty.StringVal(ref.Hash().String())
	}

	attrs[GitBlockArgumentBranch] = branch
	attrs[GitBlockArgumentTag] = tag
	attrs[GitBlockAttrIsBranch] = isBranch
	attrs[GitBlockAttrRef] = refString
	attrs[GitBlockAttrIsTag] = isTag
	attrs[GitBlockAttrIsRemote] = isRemote
	attrs[GitBlockAttrIsNote] = isNote
	attrs[GitBlockAttrSha] = sha
	attrs[GitBlockAttrLastTag] = cty.StringVal(lastTag)
	attrs[GitBlockAttrCommitsSinceLastTag] = commitsSinceLastTag

	// get the commit
	return attrs, diags
}

func parseNetrcFile(log *logrus.Entry, path string, u *url.URL) transport.AuthMethod {
	logger := log.WithField("auth", "netrc")
	n, err := netrc.Parse(path)
	if err != nil {
		logger.Warn("failed to parse netrc file: ", err)
		return nil
	}
	username := n.Machine(u.Host).Get("login")
	password := n.Machine(u.Host).Get("password")
	authMethod := &http.BasicAuth{
		Username: username,
		Password: password,
	}
	return authMethod
}

func parseGitCredentialsFile(log *logrus.Entry, path string, u *url.URL) transport.AuthMethod {
	logger := log.WithField("auth", "git-credentials")
	data, err := os.ReadFile(path)
	if err != nil {
		logger.Warn("failed to read git-credentials file: ", err)
		return nil
	}

	data = bytes.TrimSpace(data)
	lines := bytes.Split(data, []byte("\n"))
	for _, lineRaw := range lines {
		lineRaw = bytes.TrimSpace(lineRaw)
		line := string(lineRaw)
		if strings.HasPrefix(line, "#") {
			logger.Trace("skipping comment: ", line)
			continue
		}
		credentialUrl, err := url.Parse(line)
		if err != nil {
			logger.Warn("failed to parse git-credentials line: ", err)
			continue
		}

		if credentialUrl.Host == u.Host && credentialUrl.Scheme == u.Scheme && strings.HasPrefix(u.Path, credentialUrl.Path) {
			username := credentialUrl.User.Username()
			password, ok := credentialUrl.User.Password()
			if !ok {
				logger.Warn("failed to retrieve password from git-credentials file, falling back to '': ", line)
			}
			authMethod := &http.BasicAuth{
				Username: username,
				Password: password,
			}
			return authMethod
		}
	}
	return nil

}

// GetLatestTagFromRepository returns the latest tag from a git repository
// https://github.com/src-d/go-git/issues/1030#issuecomment-443679681
func GetLatestTagFromRepository(repository *git.Repository) (*plumbing.Reference, *object.Commit, error) {
	tagRefs, err := repository.Tags()
	if err != nil {
		return nil, nil, err
	}
	var commit *object.Commit
	var latestTagCommit *object.Commit
	var latestTagName *plumbing.Reference
	err = tagRefs.ForEach(func(tagRef *plumbing.Reference) error {
		revision := plumbing.Revision(tagRef.Name().String())
		tagCommitHash, err := repository.ResolveRevision(revision)
		if err != nil {
			return err
		}

		commit, err = repository.CommitObject(*tagCommitHash)
		if err != nil {
			return err
		}

		if latestTagCommit == nil {
			latestTagCommit = commit
			latestTagName = tagRef
		}

		if commit.Committer.When.After(latestTagCommit.Committer.When) {
			latestTagCommit = commit
			latestTagName = tagRef
		}

		return nil
	})
	if err != nil {
		return nil, commit, err
	}

	return latestTagName, commit, nil
}
