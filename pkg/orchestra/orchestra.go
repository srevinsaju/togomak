package orchestra

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/hashicorp/hcl/v2"
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
	"github.com/srevinsaju/togomak/v1/pkg/x"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/function/stdlib"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"time"
)

func NewLogger(cfg Config) *logrus.Logger {
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: false,
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
	logger.Infof("%s (version=%s)", meta.AppName, meta.AppVersion)

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

	// --> set up HCL context
	hclContext := &hcl.EvalContext{
		Functions: map[string]function.Function{
			"split":      stdlib.SplitFunc,
			"join":       stdlib.JoinFunc,
			"lower":      stdlib.LowerFunc,
			"upper":      stdlib.UpperFunc,
			"trim":       stdlib.TrimFunc,
			"replace":    stdlib.ReplaceFunc,
			"contains":   stdlib.ContainsFunc,
			"regex":      stdlib.RegexFunc,
			"regexall":   stdlib.RegexAllFunc,
			"max":        stdlib.MaxFunc,
			"min":        stdlib.MinFunc,
			"ceil":       stdlib.CeilFunc,
			"floor":      stdlib.FloorFunc,
			"abs":        stdlib.AbsoluteFunc,
			"format":     stdlib.FormatFunc,
			"jsonencode": stdlib.JSONEncodeFunc,
			"jsondecode": stdlib.JSONDecodeFunc,
			"timeadd":    stdlib.TimeAddFunc,

			"trimprefix": stdlib.TrimPrefixFunc,
			"trimsuffix": stdlib.TrimSuffixFunc,
			"coalesce":   stdlib.CoalesceFunc,
			"title":      stdlib.TitleFunc,
			"hasindex":   stdlib.HasIndexFunc,

			"length": stdlib.LengthFunc,
			"len":    stdlib.LengthFunc,

			"keys":       stdlib.KeysFunc,
			"values":     stdlib.ValuesFunc,
			"merge":      stdlib.MergeFunc,
			"setproduct": stdlib.SetProductFunc,
			"setunion":   stdlib.SetUnionFunc,

			"flatten": stdlib.FlattenFunc,
			"file": function.New(&function.Spec{
				Params: []function.Parameter{
					{Name: "path", Type: cty.String},
				},
				Type: function.StaticReturnType(cty.String),
				Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
					f, err := os.OpenFile(args[0].AsString(), os.O_RDONLY, 0644)
					if err != nil {
						return cty.NilVal, err
					}
					defer f.Close()
					data, err := io.ReadAll(f)
					if err != nil {
						return cty.NilVal, err
					}
					return cty.StringVal(string(data)), nil
				},
			}),
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
				Detail:      fmt.Sprintf("stage.%s can only use a single macro", stage.Id),
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
				Detail:      fmt.Sprintf("stage.%s uses an invalid macro", stage.Id),
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

		err = mergo.Merge(&stage, macro.Stage, mergo.WithOverride)

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

	for _, layer := range depGraph.TopoSortedLayers() {

		for _, runnableId := range layer {
			var runnable ci.Runnable
			var ok bool

			if runnableId == meta.RootStage {
				continue
			}
			runnable, diags = ci.Resolve(ctx, pipe, runnableId)
			if diags.HasErrors() {
				logger.Fatal(diags.Error())
				return
			}
			logger.Debugf("runnable %s is %T", runnableId, runnable)

			ok, diags = runnable.CanRun(ctx)
			if diags.HasErrors() {
				logger.Fatal(diags.Error())
			}
			runnable.Prepare(ctx, !ok)
			if !ok {
				logger.Debugf("skipping runnable %s, condition evaluated to false", runnableId)
				continue
			}

			// TODO: implement daemon kinds here
			wg.Add(1)

			go func(runnableId string) {
				stageDiags := runnable.Run(ctx)
				if !stageDiags.HasErrors() {
					wg.Done()
					return
				}
				logger.Warn(stageDiags.Error())
				logger.Infof("retrying runnable %s", runnableId)
				if runnable.CanRetry() {
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
				wg.Done()
			}(runnableId)

			if cfg.Pipeline.DryRun {
				// TODO: implement --concurrency option
				// wait for the runnable to finish
				// disable concurrency
				wg.Wait()
			}
		}
		wg.Wait()

		if diags.HasErrors() {
			diags.Fatal(logger.WriterLevel(logrus.ErrorLevel))
		}

	}

}
