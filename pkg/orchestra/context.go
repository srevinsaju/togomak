package orchestra

import (
	"context"
	"github.com/google/uuid"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/ext/tryfunc"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/sirupsen/logrus"
	"github.com/srevinsaju/togomak/v1/pkg/c"
	"github.com/srevinsaju/togomak/v1/pkg/meta"
	"github.com/srevinsaju/togomak/v1/pkg/third-party/hashicorp/terraform/lang/funcs"
	"github.com/srevinsaju/togomak/v1/pkg/x"
	ctyyaml "github.com/zclconf/go-cty-yaml"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/function/stdlib"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

type Togomak struct {
	logger        *logrus.Logger
	pipelineId    string
	cfg           Config
	cwd           string
	ectx          *hcl.EvalContext
	tempDir       string
	parser        *hclparse.Parser
	hclDiagWriter hcl.DiagnosticWriter
}

func (t Togomak) Logger() *logrus.Logger {
	return t.logger
}

func (t Togomak) Parser() *hclparse.Parser {
	return t.parser
}

func NewContextWithTogomak(cfg Config) (Togomak, context.Context) {

	logger := NewLogger(cfg)
	if !cfg.Child {
		logger.Infof("%s (version=%s)", meta.AppName, meta.AppVersion)
	}

	// --> set up the working directory
	cwd := Chdir(cfg, logger)
	// create temporary directory
	pipelineId := uuid.New().String()
	tmpDir := filepath.Join(meta.BuildDirPrefix, "pipelines", "tmp")
	err := os.MkdirAll(tmpDir, 0755)
	x.Must(err)
	tmpDir, err = os.MkdirTemp(tmpDir, pipelineId)
	x.Must(err)

	ctx := context.Background()
	ctx = context.WithValue(ctx, c.TogomakContextTempDir, tmpDir)
	ctx = context.WithValue(ctx, c.TogomakContextCi, cfg.Ci)
	ctx = context.WithValue(ctx, c.TogomakContextUnattended, cfg.Unattended)
	ctx = context.WithValue(ctx, c.TogomakContextLogger, logger)
	ctx = context.WithValue(ctx, c.TogomakContextBootTime, time.Now())
	ctx = context.WithValue(ctx, c.TogomakContextPipelineId, pipelineId)
	ctx = context.WithValue(ctx, c.TogomakContextOwd, cfg.Owd)
	ctx = context.WithValue(ctx, c.TogomakContextCwd, cwd)
	ctx = context.WithValue(ctx, c.TogomakContextHostname, cfg.Hostname)
	ctx = context.WithValue(ctx, c.TogomakContextUsername, cfg.User)
	ctx = context.WithValue(ctx, c.TogomakContextPipelineFilePath, cfg.Pipeline.FilePath)
	ctx = context.WithValue(ctx, c.TogomakContextPipelineDryRun, cfg.Pipeline.DryRun)
	ctx = context.WithValue(ctx, c.TogomakContextPipelineTmpDir, tmpDir)

	ctx = context.WithValue(ctx, c.TogomakContextMutexStages, &sync.Mutex{})
	ctx = context.WithValue(ctx, c.TogomakContextMutexData, &sync.Mutex{})
	ctx = context.WithValue(ctx, c.TogomakContextMutexLocals, &sync.Mutex{})
	ctx = context.WithValue(ctx, c.TogomakContextMutexMacro, &sync.Mutex{})

	// --> set up HCL context
	hclContext := &hcl.EvalContext{
		Functions: map[string]function.Function{
			"abs":              stdlib.AbsoluteFunc,
			"abspath":          funcs.AbsPathFunc,
			"alltrue":          funcs.AllTrueFunc,
			"anytrue":          funcs.AnyTrueFunc,
			"basename":         funcs.BasenameFunc,
			"base64decode":     funcs.Base64DecodeFunc,
			"base64encode":     funcs.Base64EncodeFunc,
			"base64gzip":       funcs.Base64GzipFunc,
			"base64sha256":     funcs.Base64Sha256Func,
			"base64sha512":     funcs.Base64Sha512Func,
			"bcrypt":           funcs.BcryptFunc,
			"can":              tryfunc.CanFunc,
			"ceil":             stdlib.CeilFunc,
			"chomp":            stdlib.ChompFunc,
			"coalesce":         funcs.CoalesceFunc,
			"coalescelist":     stdlib.CoalesceListFunc,
			"compact":          stdlib.CompactFunc,
			"concat":           stdlib.ConcatFunc,
			"contains":         stdlib.ContainsFunc,
			"csvdecode":        stdlib.CSVDecodeFunc,
			"dirname":          funcs.DirnameFunc,
			"distinct":         stdlib.DistinctFunc,
			"element":          stdlib.ElementFunc,
			"endswith":         funcs.EndsWithFunc,
			"chunklist":        stdlib.ChunklistFunc,
			"file":             funcs.MakeFileFunc(cwd, false),
			"fileexists":       funcs.MakeFileExistsFunc(cwd),
			"fileset":          funcs.MakeFileSetFunc(cwd),
			"filebase64":       funcs.MakeFileFunc(cwd, true),
			"filebase64sha256": funcs.MakeFileBase64Sha256Func(cwd),
			"filebase64sha512": funcs.MakeFileBase64Sha512Func(cwd),
			"filemd5":          funcs.MakeFileMd5Func(cwd),
			"filesha1":         funcs.MakeFileSha1Func(cwd),
			"filesha256":       funcs.MakeFileSha256Func(cwd),
			"filesha512":       funcs.MakeFileSha512Func(cwd),
			"flatten":          stdlib.FlattenFunc,
			"floor":            stdlib.FloorFunc,
			"format":           stdlib.FormatFunc,
			"formatdate":       stdlib.FormatDateFunc,
			"formatlist":       stdlib.FormatListFunc,
			"indent":           stdlib.IndentFunc,
			"index":            funcs.IndexFunc, // stdlib.IndexFunc is not compatible
			"join":             stdlib.JoinFunc,
			"jsondecode":       stdlib.JSONDecodeFunc,
			"jsonencode":       stdlib.JSONEncodeFunc,
			"keys":             stdlib.KeysFunc,
			"length":           funcs.LengthFunc,
			"list":             funcs.ListFunc,
			"log":              stdlib.LogFunc,
			"lookup":           funcs.LookupFunc,
			"lower":            stdlib.LowerFunc,
			"map":              funcs.MapFunc,
			"matchkeys":        funcs.MatchkeysFunc,
			"max":              stdlib.MaxFunc,
			"md5":              funcs.Md5Func,
			"merge":            stdlib.MergeFunc,
			"min":              stdlib.MinFunc,
			"one":              funcs.OneFunc,
			"parseint":         stdlib.ParseIntFunc,
			"pathexpand":       funcs.PathExpandFunc,
			"pow":              stdlib.PowFunc,
			"range":            stdlib.RangeFunc,
			"regex":            stdlib.RegexFunc,
			"regexall":         stdlib.RegexAllFunc,
			"replace":          funcs.ReplaceFunc,
			"reverse":          stdlib.ReverseListFunc,
			"rsadecrypt":       funcs.RsaDecryptFunc,
			"sensitive":        funcs.SensitiveFunc,
			"nonsensitive":     funcs.NonsensitiveFunc,
			"setintersection":  stdlib.SetIntersectionFunc,
			"setproduct":       stdlib.SetProductFunc,
			"setsubtract":      stdlib.SetSubtractFunc,
			"setunion":         stdlib.SetUnionFunc,
			"sha1":             funcs.Sha1Func,
			"sha256":           funcs.Sha256Func,
			"sha512":           funcs.Sha512Func,
			"signum":           stdlib.SignumFunc,
			"slice":            stdlib.SliceFunc,
			"sort":             stdlib.SortFunc,
			"split":            stdlib.SplitFunc,
			"startswith":       funcs.StartsWithFunc,
			"strcontains":      funcs.StrContainsFunc,
			"strrev":           stdlib.ReverseFunc,
			"substr":           stdlib.SubstrFunc,
			"sum":              funcs.SumFunc,
			"textdecodebase64": funcs.TextDecodeBase64Func,
			"textencodebase64": funcs.TextEncodeBase64Func,
			"timestamp":        funcs.TimestampFunc,
			"timeadd":          stdlib.TimeAddFunc,
			"timecmp":          funcs.TimeCmpFunc,
			"title":            stdlib.TitleFunc,
			"tostring":         funcs.MakeToFunc(cty.String),
			"tonumber":         funcs.MakeToFunc(cty.Number),
			"tobool":           funcs.MakeToFunc(cty.Bool),
			"toset":            funcs.MakeToFunc(cty.Set(cty.DynamicPseudoType)),
			"tolist":           funcs.MakeToFunc(cty.List(cty.DynamicPseudoType)),
			"tomap":            funcs.MakeToFunc(cty.Map(cty.DynamicPseudoType)),
			"transpose":        funcs.TransposeFunc,
			"trim":             stdlib.TrimFunc,
			"trimprefix":       stdlib.TrimPrefixFunc,
			"trimspace":        stdlib.TrimSpaceFunc,
			"trimsuffix":       stdlib.TrimSuffixFunc,
			"try":              tryfunc.TryFunc,
			"upper":            stdlib.UpperFunc,
			"urlencode":        funcs.URLEncodeFunc,
			"uuid":             funcs.UUIDFunc,
			"uuidv5":           funcs.UUIDV5Func,
			"values":           stdlib.ValuesFunc,
			"which": function.New(&function.Spec{
				Params: []function.Parameter{
					{
						Name:             "executable",
						AllowDynamicType: true,
						Type:             cty.String,
					},
				},
				Type: function.StaticReturnType(cty.String),
				Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
					path, err := exec.LookPath(args[0].AsString())
					if err != nil {
						return cty.StringVal(""), err
					}
					return cty.StringVal(path), nil
				},
				Description: "Returns the absolute path to an executable in the current PATH.",
			}),
			"yamldecode": ctyyaml.YAMLDecodeFunc,
			"yamlencode": ctyyaml.YAMLEncodeFunc,
			"zipmap":     stdlib.ZipmapFunc,
		},

		Variables: map[string]cty.Value{
			"true":  cty.True,
			"false": cty.False,
			"null":  cty.NullVal(cty.DynamicPseudoType),

			c.TogomakContextOwd:      cty.StringVal(cfg.Owd),
			c.TogomakContextCwd:      cty.StringVal(cwd),
			c.TogomakContextHostname: cty.StringVal(cfg.Hostname),
			c.TogomakContextUsername: cty.StringVal(cfg.User),

			"pipeline": cty.ObjectVal(map[string]cty.Value{
				"id":     cty.StringVal(pipelineId),
				"path":   cty.StringVal(cfg.Pipeline.FilePath),
				"tmpDir": cty.StringVal(tmpDir),
			}),

			"togomak": cty.ObjectVal(map[string]cty.Value{
				"version":        cty.StringVal(meta.AppVersion),
				"boot_time":      cty.StringVal(time.Now().Format(time.RFC3339)),
				"boot_time_unix": cty.NumberIntVal(time.Now().Unix()),
				"pipeline_id":    cty.StringVal(pipelineId),
				"ci":             cty.BoolVal(cfg.Ci),
				"unattended":     cty.BoolVal(cfg.Unattended),
			}),
		},
	}

	parser := hclparse.NewParser()
	diagnosticTextWriter := hcl.NewDiagnosticTextWriter(logger.Writer(), parser.Files(), 0, true)
	ctx = context.WithValue(ctx, c.TogomakContextHclDiagWriter, diagnosticTextWriter)

	ctx = context.WithValue(ctx, c.TogomakContextHclEval, hclContext)
	t := Togomak{
		logger:        logger,
		pipelineId:    pipelineId,
		cfg:           cfg,
		cwd:           cwd,
		hclDiagWriter: diagnosticTextWriter,
		parser:        parser,
		ectx:          hclContext,
		tempDir:       tmpDir,
	}

	return t, ctx
}
