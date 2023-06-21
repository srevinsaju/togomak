package data

import (
	"context"
	"fmt"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/go-git/go-git/v5/storage"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/hashicorp/hcl/v2"
	"github.com/sirupsen/logrus"
	"github.com/srevinsaju/togomak/v1/pkg/c"
	"github.com/srevinsaju/togomak/v1/pkg/diag"
	"github.com/srevinsaju/togomak/v1/pkg/ui"
	"github.com/zclconf/go-cty/cty"
	"io"
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

func (e *GitProvider) DecodeBody(body hcl.Body) diag.Diagnostics {
	if !e.initialized {
		panic("provider not initialized")
	}
	var diags diag.Diagnostics
	hclDiagWriter := e.ctx.Value(c.TogomakContextHclDiagWriter).(hcl.DiagnosticWriter)
	hclContext := e.ctx.Value(c.TogomakContextHclEval).(*hcl.EvalContext)

	schema := e.Schema()
	content, hclDiags := body.Content(schema)
	if hclDiags.HasErrors() {
		diags = diags.NewHclWriteDiagnosticsError(e.Identifier(), hclDiagWriter.WriteDiagnostics(hclDiags))
	}

	repo, d := content.Attributes[GitBlockArgumentUrl].Expr.Value(hclContext)
	hclDiags = append(hclDiags, d...)

	tagAttr, ok := content.Attributes[GitBlockArgumentTag]
	tag := cty.StringVal("")
	if ok {
		tag, d = tagAttr.Expr.Value(hclContext)
		hclDiags = append(hclDiags, d...)
	}

	branchAttr, ok := content.Attributes[GitBlockArgumentBranch]
	branch := cty.StringVal("")
	if ok {
		branch, d = branchAttr.Expr.Value(hclContext)
		hclDiags = append(hclDiags, d...)
	}

	commitAttr, ok := content.Attributes[GitBlockArgumentCommit]
	commit := cty.StringVal("")
	if ok {
		commit, d = commitAttr.Expr.Value(hclContext)
		hclDiags = append(hclDiags, d...)
	}

	destinationAttr, ok := content.Attributes[GitBlockArgumentDestination]
	destination := cty.StringVal("")
	if ok {
		destination, d = destinationAttr.Expr.Value(hclContext)
		hclDiags = append(hclDiags, d...)
	}

	depthAttr, ok := content.Attributes[GitBlockArgumentDepth]
	depth := cty.NumberIntVal(0)
	if ok {
		depth, d = depthAttr.Expr.Value(hclContext)
		hclDiags = append(hclDiags, d...)
	}

	caBundleAttr, ok := content.Attributes[GitBlockArgumentCaBundle]
	caBundle := cty.StringVal("")
	if ok {
		caBundle, d = caBundleAttr.Expr.Value(hclContext)
		hclDiags = append(hclDiags, d...)
	}

	filesAttr, ok := content.Attributes[GitBlockArgumentFiles]
	files := cty.ListValEmpty(cty.String)
	if ok {
		files, d = filesAttr.Expr.Value(hclContext)
		hclDiags = append(hclDiags, d...)
	}

	authBlock := content.Blocks.OfType(GitBlockArgumentAuth)
	var authConfig gitProviderAuthConfig
	if len(authBlock) == 1 {
		auth, d := content.Blocks.OfType(GitBlockArgumentAuth)[0].Body.Content(GitProviderAuthSchema())
		hclDiags = append(hclDiags, d...)

		authUsername, d := auth.Attributes["username"].Expr.Value(hclContext)
		hclDiags = append(hclDiags, d...)

		authPassword, d := auth.Attributes["password"].Expr.Value(hclContext)
		hclDiags = append(hclDiags, d...)

		authSshPassword, d := auth.Attributes["ssh_password"].Expr.Value(hclContext)
		hclDiags = append(hclDiags, d...)

		authSshPrivateKey, d := auth.Attributes["ssh_private_key"].Expr.Value(hclContext)
		hclDiags = append(hclDiags, d...)

		authConfig = gitProviderAuthConfig{
			username:      authUsername.AsString(),
			password:      authPassword.AsString(),
			sshPassword:   authSshPassword.AsString(),
			sshPrivateKey: authSshPrivateKey.AsString(),
		}
	}

	if hclDiags.HasErrors() {
		diags = diags.NewHclWriteDiagnosticsError(e.Identifier(), hclDiagWriter.WriteDiagnostics(hclDiags))
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

	return nil
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

func (e *GitProvider) Value(ctx context.Context, id string) string {
	if !e.initialized {
		panic("provider not initialized")
	}
	return ""
}

func (e *GitProvider) Attributes(ctx context.Context) map[string]cty.Value {
	logger := ctx.Value(c.TogomakContextLogger).(*logrus.Logger).WithField("provider", e.Name())
	var diags diag.Diagnostics
	if !e.initialized {
		panic("provider not initialized")
	}

	var attrs = make(map[string]cty.Value)

	// clone git repo
	// git clone
	fmt.Println(e.cfg.repo)
	var s storage.Storer
	var authMethod transport.AuthMethod
	if e.cfg.auth.password != "" {
		authMethod = &http.BasicAuth{
			Username: e.cfg.auth.username,
			Password: e.cfg.auth.password,
		}
	} else if e.cfg.auth.isSsh {
		publicKeys, err := ssh.NewPublicKeys(e.cfg.auth.username, []byte(e.cfg.auth.sshPrivateKey), e.cfg.auth.sshPassword)
		if err != nil {
			diags = diags.Append(diag.Diagnostic{
				Severity: diag.SeverityError,
				Summary:  "ssh key error",
				Detail:   err.Error(),
			})
			return nil
		}
		authMethod = publicKeys
	}

	cloneOptions := &git.CloneOptions{
		Tags:     git.AllTags,
		Depth:    e.cfg.depth,
		CABundle: e.cfg.caBundle,
		Auth:     authMethod,
		URL:      e.cfg.repo,
		Progress: nil,
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
		diags = diags.Append(diag.Diagnostic{
			Severity: diag.SeverityError,
			Summary:  "git clone failed",
			Detail:   err.Error(),
			Source:   e.Identifier(),
		})
		diags.Write(logger.WriterLevel(logrus.ErrorLevel))
		return nil
	}

	w, err := repo.Worktree()
	if err != nil {
		diags = diags.Append(diag.Diagnostic{
			Severity: diag.SeverityError,
			Summary:  "checkout failed",
			Detail:   err.Error(),
			Source:   e.Identifier(),
		})
		diags.Write(logger.WriterLevel(logrus.ErrorLevel))
		return nil
	}

	commitIter, err := repo.Log(&git.LogOptions{
		Order: git.LogOrderCommitterTime,
	})

	count := 0
	lastTag := ""
	if err != nil {
		diags = diags.Append(diag.Diagnostic{
			Severity: diag.SeverityWarning,
			Summary:  "git log failed",
			Detail:   err.Error(),
			Source:   e.Identifier(),
		})
	} else {
		for commit, err := commitIter.Next(); err != nil; {
			if commit.Type() == plumbing.TagObject {
				lastTag = commit.Hash.String()
			}
			count++
		}
	}
	commitsSinceLastTag := cty.NumberIntVal(int64(count))

	var files = make(map[string]cty.Value)
	for _, f := range e.cfg.files {
		_, err := w.Filesystem.Stat(f)
		if err != nil {
			diags = diags.Append(diag.Diagnostic{
				Severity: diag.SeverityError,
				Summary:  "git file search failed",
				Detail:   err.Error(),
				Source:   e.Identifier(),
			})
			continue
		}
		file, err := w.Filesystem.Open(f)
		if err != nil {
			diags = diags.Append(diag.Diagnostic{
				Severity: diag.SeverityError,
				Summary:  "git file open failed",
				Detail:   err.Error(),
				Source:   e.Identifier(),
			})
			continue
		}
		defer file.Close()

		data, err := io.ReadAll(file)
		if err != nil {
			diags = diags.Append(diag.Diagnostic{
				Severity: diag.SeverityError,
				Summary:  "git file read failed",
				Detail:   err.Error(),
				Source:   e.Identifier(),
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
		diags = diags.Append(diag.Diagnostic{
			Severity: diag.SeverityWarning,
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
	return attrs
}
