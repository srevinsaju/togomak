package data

import (
	"context"
	"fmt"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
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

	repo, d := content.Attributes["url"].Expr.Value(hclContext)
	hclDiags = append(hclDiags, d...)

	tagAttr, ok := content.Attributes["tag"]
	tag := cty.StringVal("")
	if ok {
		tag, d = tagAttr.Expr.Value(hclContext)
		hclDiags = append(hclDiags, d...)
	}

	branchAttr, ok := content.Attributes["branch"]
	branch := cty.StringVal("")
	if ok {
		branch, d = branchAttr.Expr.Value(hclContext)
		hclDiags = append(hclDiags, d...)
	}

	commitAttr, ok := content.Attributes["commit"]
	commit := cty.StringVal("")
	if ok {
		commit, d = commitAttr.Expr.Value(hclContext)
		hclDiags = append(hclDiags, d...)
	}

	destinationAttr, ok := content.Attributes["destination"]
	destination := cty.StringVal("")
	if ok {
		destination, d = destinationAttr.Expr.Value(hclContext)
		hclDiags = append(hclDiags, d...)
	}

	depthAttr, ok := content.Attributes["dreaderepth"]
	depth := cty.NumberIntVal(0)
	if ok {
		depth, d = depthAttr.Expr.Value(hclContext)
		hclDiags = append(hclDiags, d...)
	}

	caBundleAttr, ok := content.Attributes["ca_bundle"]
	caBundle := cty.StringVal("")
	if ok {
		caBundle, d = caBundleAttr.Expr.Value(hclContext)
		hclDiags = append(hclDiags, d...)
	}

	filesAttr, ok := content.Attributes["files"]
	files := cty.ListValEmpty(cty.String)
	if ok {
		files, d = filesAttr.Expr.Value(hclContext)
		hclDiags = append(hclDiags, d...)
	}

	authBlock := content.Blocks.OfType("auth")
	var authConfig gitProviderAuthConfig
	if len(authBlock) == 1 {
		auth, d := content.Blocks.OfType("auth")[0].Body.Content(GitProviderAuthSchema())
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
				Type: "auth",
			},
		},
		Attributes: []hcl.AttributeSchema{
			{
				Name:     "url",
				Required: true,
			},
			{
				Name:     "tag",
				Required: false,
			},
			{
				Name:     "branch",
				Required: false,
			},
			{
				Name:     "commit",
				Required: false,
			},
			{
				Name:     "destination",
				Required: false,
			},
			{
				Name:     "depth",
				Required: false,
			},
			{
				Name:     "files",
				Required: false,
			},
		},
	}
}

func (e *GitProvider) Initialized() bool {
	return e.initialized
}

func (e *GitProvider) Value(ctx context.Context) string {
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

	attrs["files"] = cty.MapVal(files)
	attrs["url"] = cty.StringVal(e.cfg.repo)
	attrs["tag"] = cty.StringVal(e.cfg.tag)

	ref, err := repo.Head()
	branch := cty.StringVal("")
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
		branch = cty.StringVal(ref.Name().String())
		isBranch = cty.BoolVal(ref.Name().IsBranch())
		isTag = cty.BoolVal(ref.Name().IsTag())
		isRemote = cty.BoolVal(ref.Name().IsRemote())
		isNote = cty.BoolVal(ref.Name().IsNote())
		sha = cty.StringVal(ref.Hash().String())
	}
	attrs["branch"] = branch
	attrs["is_branch"] = isBranch
	attrs["is_tag"] = isTag
	attrs["is_remote"] = isRemote
	attrs["is_note"] = isNote
	attrs["sha"] = sha

	// get the commit
	return attrs
}
