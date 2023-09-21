package ci

import (
	"context"
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/srevinsaju/togomak/v1/pkg/runnable"
	"github.com/srevinsaju/togomak/v1/pkg/x"
	"strings"

	"sync"
)

const ThisBlock = "this"

type Retryable interface {
	// CanRetry decides if the runnable can be retried
	CanRetry() bool
	// MaxRetries returns the maximum number of retries that is valid for
	// this runnable
	MaxRetries() int
	// MinRetryBackoff returns the minimum backoff time in seconds
	MinRetryBackoff() int
	// MaxRetryBackoff returns the maximum backoff time in seconds
	MaxRetryBackoff() int
	// RetryExponentialBackoff returns true if the backoff time should be
	// exponentially increasing
	RetryExponentialBackoff() bool
}

type Describable interface {
	Description() string
	Identifier() string
	Type() string
}

type Runnable interface {
	// Prepare is called before the runnable is run
	Prepare(ctx context.Context, skip bool, overridden bool) hcl.Diagnostics
	Run(ctx context.Context, options ...runnable.Option) (diags hcl.Diagnostics)
	CanRun(ctx context.Context, options ...runnable.Option) (ok bool, diags hcl.Diagnostics)
}

type RunnableWithHooks interface {
	Runnable
	BeforeRun(ctx context.Context, opts ...runnable.Option) hcl.Diagnostics
	AfterRun(ctx context.Context, opts ...runnable.Option) hcl.Diagnostics
}

type Traversable interface {
	Variables() []hcl.Traversal
}

type Contextual interface {
	Set(k any, v any)
	Get(k any) any
}

type Killable interface {
	Terminate(safe bool) hcl.Diagnostics
	Kill() hcl.Diagnostics
	Terminated() bool
}

type Daemon interface {
	// IsDaemon returns true if the runnable is a daemon
	IsDaemon() bool
	Lifecycle(ctx context.Context) (*DaemonLifecycle, hcl.Diagnostics)
}

type Block interface {
	Retryable
	Describable
	Contextual
	Traversable
	Runnable
	Killable
	Daemon
}

type Blocks []Block

func (r Blocks) Variables() []hcl.Traversal {
	var traversal []hcl.Traversal
	for _, runnable := range r {
		traversal = append(traversal, runnable.Variables()...)
	}
	return traversal
}

func (r Blocks) Run(ctx context.Context, opts ...runnable.Option) hcl.Diagnostics {
	// run all runnables in parallel, collect errors and return
	// create a channel to receive errors
	var wg sync.WaitGroup
	errChan := make(chan error, len(r))
	for _, runnable := range r {
		wg.Add(1)
		go func(runnable Block) {
			defer wg.Done()
			errChan <- runnable.Run(ctx, opts...)
		}(runnable)
	}
	wg.Wait()
	close(errChan)

	return nil
}

func Resolve(pipe *Pipeline, id string) (Block, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	blocks := strings.Split(id, ".")
	if len(blocks) != 2 && len(blocks) != 3 {
		diags = diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid identifier",
			Detail:   fmt.Sprintf("Expected a valid identifier, got '%s'", id),
		})
	}
	if diags.HasErrors() {
		return nil, diags
	}

	switch blocks[0] {
	case "provider":
		return nil, diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Unsupported identifier",
			Detail:   fmt.Sprintf("Expected a valid identifier, got %s", id),
		})
	case StageBlock:
		stage, d := pipe.Stages.ById(blocks[1])
		diags = diags.Extend(d)
		return stage, diags
	case DataBlock:
		data, d := pipe.Data.ById(blocks[1], blocks[2])
		diags = diags.Extend(d)
		return data, diags
	case MacroBlock:
		macro, d := pipe.Macros.ById(blocks[1])
		diags = diags.Extend(d)
		return macro, diags
	case LocalBlock:
		local, d := pipe.Local.ById(blocks[1])
		diags = diags.Extend(d)
		return local, diags
	case LocalsBlock:
		panic("locals block cannot be resolved")
	case ModuleBlock:
		module, d := pipe.Modules.ById(blocks[1])
		diags = diags.Extend(d)
		return module, diags

	case ThisBlock:
		return nil, nil
	case ParamBlock:
		return nil, nil
	}

	return nil, diags.Append(&hcl.Diagnostic{
		Severity: hcl.DiagError,
		Summary:  "Unsupported identifier",
		Detail:   fmt.Sprintf("Expected a valid identifier, got %s", id),
	})
}

func ResolveFromTraversal(variable hcl.Traversal) (string, hcl.Diagnostics) {
	blockType := variable.RootName()
	var parent string
	var diags hcl.Diagnostics
	switch blockType {
	case DataBlock:
		// the data block has the provider type as well as the name
		provider := variable[1].(hcl.TraverseAttr).Name
		name := variable[2].(hcl.TraverseAttr).Name
		parent = x.RenderBlock(DataBlock, provider, name)
	case StageBlock:
		// the stage block has the name
		name := variable[1].(hcl.TraverseAttr).Name
		parent = x.RenderBlock(StageBlock, name)
	case LocalBlock:
		// the local block has the name
		name := variable[1].(hcl.TraverseAttr).Name
		parent = x.RenderBlock(LocalBlock, name)
	case MacroBlock:
		// the local block has the name
		name := variable[1].(hcl.TraverseAttr).Name
		parent = x.RenderBlock(MacroBlock, name)
	case ModuleBlock:
		// the module block has the name
		name := variable[1].(hcl.TraverseAttr).Name
		parent = x.RenderBlock(ModuleBlock, name)
	case ParamBlock, ThisBlock, BuilderBlock:
		return "", nil
	default:
		return "", nil

	}
	return parent, diags

}
