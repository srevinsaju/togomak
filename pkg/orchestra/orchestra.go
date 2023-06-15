package orchestra

import (
	"context"
	"fmt"
	"github.com/alessio/shellescape"
	"github.com/google/uuid"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/ext/tryfunc"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/imdario/mergo"
	"github.com/kendru/darwin/go/depgraph"
	"github.com/sirupsen/logrus"
	"github.com/srevinsaju/togomak/v1/pkg/c"
	"github.com/srevinsaju/togomak/v1/pkg/ci"
	"github.com/srevinsaju/togomak/v1/pkg/diag"
	"github.com/srevinsaju/togomak/v1/pkg/graph"
	"github.com/srevinsaju/togomak/v1/pkg/meta"
	"github.com/srevinsaju/togomak/v1/pkg/pipeline"
	"github.com/srevinsaju/togomak/v1/pkg/third-party/hashicorp/terraform/lang/funcs"
	"github.com/srevinsaju/togomak/v1/pkg/x"
	ctyyaml "github.com/zclconf/go-cty-yaml"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/function/stdlib"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

func NewLogger(cfg Config) *logrus.Logger {
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:    false,
		DisableTimestamp: cfg.Child,
	})
	switch cfg.Verbosity {
	case -1:
	case 0:
		logger.SetLevel(logrus.InfoLevel)
		break
	case 1:
		logger.SetLevel(logrus.DebugLevel)
		break
	case 2:
	default:
		logger.SetLevel(logrus.TraceLevel)
		break
	}
	if cfg.Child {
		logger.SetFormatter(&logrus.TextFormatter{
			DisableTimestamp:          true,
			DisableColors:             false,
			EnvironmentOverrideColors: false,
			ForceColors:               true,
			ForceQuote:                false,
		})
	}
	return logger
}

func Chdir(cfg Config, logger *logrus.Logger) string {
	cwd := cfg.Dir
	if cwd == "" {
		cwd = filepath.Dir(cfg.Pipeline.FilePath)
		if filepath.Base(cwd) == meta.BuildDirPrefix {
			cwd = filepath.Dir(cwd)
		}
	}
	err := os.Chdir(cwd)
	if err != nil {
		logger.Fatal(err)
	}
	return cwd
}

func Orchestra(cfg Config) {
	// --> set up the logger
	logger := NewLogger(cfg)
	if !cfg.Child {
		logger.Infof("%s (version=%s)", meta.AppName, meta.AppVersion)
	}

	// --> set up the working directory
	cwd := Chdir(cfg, logger)

	// --> set up the context
	// we will now create a long-running background context
	// and gather necessary data before reading the pipeline
	pipelineId := uuid.New().String()
	tmpDir := filepath.Join(meta.BuildDirPrefix, "pipelines", "tmp")
	err := os.MkdirAll(tmpDir, 0755)
	x.Must(err)
	tmpDir, err = os.MkdirTemp(tmpDir, pipelineId)
	x.Must(err)
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
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

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	go func() {
		select {
		case <-ch:
			logger.Warn("received interrupt signal, cancelling the pipeline")
			cancel()
		case <-ctx.Done():
			logger.Infof("took %s to complete the pipeline", time.Since(ctx.Value(c.TogomakContextBootTime).(time.Time)))
			return
		}
	}()

	// region: external parameters
	paramsGo := make(map[string]cty.Value)
	if cfg.Child {

		m := make(map[string]string)
		for _, e := range os.Environ() {
			if i := strings.Index(e, "="); i >= 0 {
				if strings.HasPrefix(e[:i], ci.TogomakParamEnvVarPrefix) {
					m[e[:i]] = e[i+1:]
				}
			}
		}
		for k, v := range m {
			if ci.TogomakParamEnvVarRegex.MatchString(k) {
				paramsGo[ci.TogomakParamEnvVarRegex.FindStringSubmatch(k)[1]] = cty.StringVal(v)
			}
		}
	}

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

			"param": cty.ObjectVal(paramsGo),

			"togomak": cty.ObjectVal(map[string]cty.Value{
				"version":        cty.StringVal(meta.AppVersion),
				"boot_time":      cty.StringVal(time.Now().Format(time.RFC3339)),
				"boot_time_unix": cty.NumberIntVal(time.Now().Unix()),
				"pipeline_id":    cty.StringVal(pipelineId),
			}),
		},
	}
	ctx = context.WithValue(ctx, c.TogomakContextHclEval, hclContext)

	// --> parse the config file
	// we will now read the pipeline from togomak.hcl
	parser := hclparse.NewParser()
	var diags diag.Diagnostics
	dgwriter := hcl.NewDiagnosticTextWriter(logger.Writer(), parser.Files(), 0, true)
	ctx = context.WithValue(ctx, c.TogomakContextHclDiagWriter, dgwriter)
	pipe, hclDiags := pipeline.Read(ctx, parser)
	if hclDiags.HasErrors() {
		logger.Fatal(dgwriter.WriteDiagnostics(hclDiags))
	}

	// whitelist all stages if unspecified
	var stageStatuses ConfigPipelineStageList = cfg.Pipeline.Stages

	// write the pipeline to the temporary directory
	pipelineFilePath := filepath.Join(tmpDir, meta.ConfigFileName)
	pipelineData := []byte{}
	for _, f := range parser.Files() {
		pipelineData = append(pipelineData, f.Bytes...)
	}
	err = os.WriteFile(pipelineFilePath, pipelineData, 0644)
	if err != nil {
		logger.Fatal(err)
	}

	/// we will first expand all local blocks
	locals, d := pipe.Locals.Expand()
	if d.HasErrors() {
		d.Fatal(logger.WriterLevel(logrus.ErrorLevel))
	}
	pipe.Local = locals

	// expand stages using macros
	for stageIdx, stage := range pipe.Stages {

		if stage.Use == nil {
			// this stage does not use a macro
			continue
		}
		v := stage.Use.Macro.Variables()
		if v == nil || len(v) == 0 {
			// this stage does not use a macro
			continue
		}
		if len(v) != 1 {
			hclDiags = hclDiags.Append(&hcl.Diagnostic{
				Severity:    hcl.DiagError,
				Summary:     "invalid macro",
				Detail:      fmt.Sprintf("%s can only use a single macro", stage.Identifier()),
				EvalContext: hclContext,
				Subject:     v[0].SourceRange().Ptr(),
			})
			logger.Fatal(dgwriter.WriteDiagnostics(hclDiags))
		}
		variable := v[0]
		if variable.RootName() != ci.MacroBlock {
			hclDiags = hclDiags.Append(&hcl.Diagnostic{
				Severity:    hcl.DiagError,
				Summary:     "invalid macro",
				Detail:      fmt.Sprintf("%s uses an invalid macro", stage.Identifier()),
				EvalContext: hclContext,
				Subject:     v[0].SourceRange().Ptr(),
			})
			logger.Fatal(dgwriter.WriteDiagnostics(hclDiags))
		}

		macroName := variable[1].(hcl.TraverseAttr).Name
		logger.Debugf("stage.%s uses macro.%s", stage.Id, macroName)
		macroRunnable, d := ci.Resolve(ctx, pipe, fmt.Sprintf("macro.%s", macroName))
		if d.HasErrors() {
			d.Fatal(logger.WriterLevel(logrus.ErrorLevel))
		}
		macro := macroRunnable.(*ci.Macro)

		oldStageId := stage.Id
		oldStageName := stage.Name
		oldStageDependsOn := stage.DependsOn

		if macro.Source != "" {
			executable, err := os.Executable()
			if err != nil {
				panic(err)
			}
			parent := shellescape.Quote(stage.Id)
			stage.Args = hcl.StaticExpr(
				cty.ListVal([]cty.Value{
					cty.StringVal(executable),
					cty.StringVal("--child"),
					cty.StringVal("--dir"), cty.StringVal(cwd),
					cty.StringVal("--file"), cty.StringVal(macro.Source),
					cty.StringVal("--parent"), cty.StringVal(parent),
				}), hcl.Range{Filename: "memory"})

		} else {
			err = mergo.Merge(&stage, macro.Stage, mergo.WithOverride)
		}

		if err != nil {
			panic(err)
		}
		stage := stage
		stage.Id = oldStageId
		stage.Name = oldStageName
		stage.DependsOn = oldStageDependsOn

		pipe.Stages[stageIdx] = stage
	}

	// --> validate the pipeline
	// TODO: validate the pipeline

	// --> generate a dependency graph
	// we will now generate a dependency graph from the pipeline
	// this will be used to generate the pipeline
	var depGraph *depgraph.Graph
	depGraph, diags = graph.TopoSort(ctx, pipe)
	if diags.HasErrors() {

		diags.Fatal(logger.WriterLevel(logrus.ErrorLevel))
	}

	// --> run the pipeline
	// we will now run the pipeline

	var diagsMutex sync.Mutex
	var wg sync.WaitGroup
	var daemonWg sync.WaitGroup
	var hasDaemons bool

	for _, layer := range depGraph.TopoSortedLayers() {
		for _, runnableId := range layer {
			var runnable ci.Runnable
			var ok bool

			if runnableId == meta.RootStage {
				continue
			}

			runnable, diags = ci.Resolve(ctx, pipe, runnableId)
			if diags.HasErrors() {
				diags.Fatal(logger.WriterLevel(logrus.ErrorLevel))
				return
			}
			logger.Debugf("runnable %s is %T", runnableId, runnable)

			ok, diags = runnable.CanRun(ctx)
			if diags.HasErrors() {
				diags.Fatal(logger.WriterLevel(logrus.ErrorLevel))
				return
			}

			// region: requested stages, whitelisting and blacklisting
			overridden := false
			if runnable.Type() == ci.StageBlock {
				stageStatus, stageStatusOk := stageStatuses.Get(runnableId)

				// when a particular stage is explicitly requested, for example
				// in the pipeline containing the following stages
				// - hello_1
				// - hello_2
				// - hello_3
				// - hello_4 (depends on hello_1)
				// if 'hello_1' is explicitly requested, we will run 'hello_4' as well
				if stageStatuses.HasOperationType(ConfigPipelineStageRunOperation) && !stageStatusOk {
					isDependentOfRequestedStage := false
					for _, ss := range stageStatuses {
						if ss.Operation == ConfigPipelineStageRunOperation {
							if depGraph.DependsOn(runnableId, ss.RunnableId()) {
								isDependentOfRequestedStage = true
								break
							}
						}
					}

					// if this stage is not dependent on the requested stage, we will skip it
					if !isDependentOfRequestedStage {
						ok = false
					}
				}

				if stageStatusOk {
					// overridden status is shown on the build pipeline if the
					// stage is explicitly whitelisted or blacklisted
					// using the ^ or + prefix
					overridden = true
					ok = ok || stageStatus.Operation == ConfigPipelineStageRunWhitelistOperation
					if stageStatus.Operation == ConfigPipelineStageRunBlacklistOperation {
						ok = false
					}
				}

			}
			// endregion: requested stages, whitelisting and blacklisting

			d := runnable.Prepare(ctx, !ok, overridden)
			if d.HasErrors() {
				diags.Fatal(logger.WriterLevel(logrus.ErrorLevel))
				return
			}

			if !ok {
				logger.Debugf("skipping runnable %s, condition evaluated to false", runnableId)
				continue
			}

			if runnable.IsDaemon() {
				hasDaemons = true
				daemonWg.Add(1)
			} else {
				wg.Add(1)
			}

			go func(runnableId string) {
				stageDiags := runnable.Run(ctx)
				if !stageDiags.HasErrors() {
					wg.Done()
					return
				}
				logger.Warn(stageDiags.Error())
				if runnable.CanRetry() {
					logger.Infof("retrying runnable %s", runnableId)
					retryCount := 0
					retryMinBackOff := time.Duration(runnable.MinRetryBackoff()) * time.Second
					retryMaxBackOff := time.Duration(runnable.MaxRetryBackoff()) * time.Second
					retrySuccess := false
					for retryCount < runnable.MaxRetries() {
						retryCount++
						sleepDuration := time.Duration(1) * time.Second
						if runnable.RetryExponentialBackoff() {

							if retryMinBackOff*time.Duration(retryCount) > retryMaxBackOff && retryMaxBackOff > 0 {
								sleepDuration = retryMaxBackOff
							} else {
								sleepDuration = retryMinBackOff * time.Duration(retryCount)
							}
						} else {
							sleepDuration = retryMinBackOff
						}
						logger.Warnf("runnable %s failed, retrying in %s", runnableId, sleepDuration)
						time.Sleep(sleepDuration)
						sDiags := runnable.Run(ctx)
						stageDiags = append(stageDiags, sDiags...)

						if !sDiags.HasErrors() {
							retrySuccess = true
							break
						}
					}

					if !retrySuccess {
						logger.Warnf("runnable %s failed after %d retries", runnableId, retryCount)
					}

				}
				diagsMutex.Lock()
				diags = diags.Extend(stageDiags)
				diagsMutex.Unlock()
				if runnable.IsDaemon() {
					daemonWg.Done()
				} else {
					wg.Done()
				}

			}(runnableId)

			if cfg.Pipeline.DryRun {
				// TODO: implement --concurrency option
				// wait for the runnable to finish
				// disable concurrency
				wg.Wait()
				daemonWg.Wait()
			}
		}
		wg.Wait()

		if diags.HasErrors() {
			diags.Write(logger.WriterLevel(logrus.ErrorLevel))
			if hasDaemons && !cfg.Pipeline.DryRun && !cfg.Unattended {
				logger.Info("pipeline failed, waiting for daemons to shut down")
				logger.Info("hit Ctrl+C to force stop them")
				// wait for daemons to stop
				daemonWg.Wait()
			} else if hasDaemons && !cfg.Pipeline.DryRun {
				logger.Info("pipeline failed, waiting for daemons to shut down...")
				// wait for daemons to stop
				cancel()
			}

		}

	}

	daemonWg.Wait()

}
