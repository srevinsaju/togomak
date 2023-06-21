package ci

import (
	"context"
	"fmt"
	"github.com/hashicorp/hcl/v2"
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
	Run(ctx context.Context) hcl.Diagnostics
	CanRun(ctx context.Context) (bool, hcl.Diagnostics)
}

type Traversable interface {
	Variables() []hcl.Traversal
}

type Contextual interface {
	Set(k any, v any)
	Get(k any) any
}

type Killable interface {
	Terminate() hcl.Diagnostics
	Kill() hcl.Diagnostics
}

type Daemon interface {
	// IsDaemon returns true if the runnable is a daemon
	IsDaemon() bool
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

func (r Blocks) Run(ctx context.Context) hcl.Diagnostics {
	// run all runnables in parallel, collect errors and return
	// create a channel to receive errors
	var wg sync.WaitGroup
	errChan := make(chan error, len(r))
	for _, runnable := range r {
		wg.Add(1)
		go func(runnable Block) {
			defer wg.Done()
			errChan <- runnable.Run(ctx)
		}(runnable)
	}
	wg.Wait()
	close(errChan)

	return nil
}

func Resolve(ctx context.Context, pipe *Pipeline, id string) (Block, hcl.Diagnostics) {
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
