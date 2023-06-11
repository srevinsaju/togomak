package ci

import (
	"context"
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/srevinsaju/togomak/v1/pkg/diag"
	"strings"

	"sync"
)

type Runnable interface {
	Name() string
	Description() string
	Identifier() string

	Run(ctx context.Context) diag.Diagnostics
	CanRun(ctx context.Context) (bool, diag.Diagnostics)

	// CanRetry decides if the runnable can be retried
	CanRetry() bool

	// Prepare is called before the runnable is run
	Prepare(ctx context.Context, skip bool)

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

	Variables() []hcl.Traversal
}

type Runnables []Runnable

func (r Runnables) Variables() []hcl.Traversal {
	var traversal []hcl.Traversal
	for _, runnable := range r {
		traversal = append(traversal, runnable.Variables()...)
	}
	return traversal
}

func (r Runnables) Run(ctx context.Context) diag.Diagnostics {
	// run all runnables in parallel, collect errors and return
	// create a channel to receive errors
	var wg sync.WaitGroup
	errChan := make(chan error, len(r))
	for _, runnable := range r {
		wg.Add(1)
		go func(runnable Runnable) {
			defer wg.Done()
			errChan <- runnable.Run(ctx)
		}(runnable)
	}
	wg.Wait()
	close(errChan)

	return nil
}

func Resolve(ctx context.Context, pipe *Pipeline, id string) (Runnable, diag.Diagnostics) {
	var diags diag.Diagnostics
	blocks := strings.Split(id, ".")
	if len(blocks) != 2 && len(blocks) != 3 {
		diags = diags.Append(diag.Diagnostic{
			Severity: diag.SeverityError,
			Summary:  "Invalid identifier",
			Detail:   fmt.Sprintf("Expected a valid identifier, got %s", id),
			Source:   "resolve",
		})
	}
	if diags.HasErrors() {
		return nil, diags
	}

	switch blocks[0] {
	case "provider":
		return nil, diags.Append(diag.NewNotImplementedError("provider"))
	case "stage":
		stage, err := pipe.Stages.ById(blocks[1])
		if err != nil {
			diags = diags.Append(diag.NewError("resolve", err.Error()))
		}
		return stage, diags
	case "data":
		data, err := pipe.Data.ById(blocks[1], blocks[2])
		if err != nil {
			diags = diags.Append(diag.NewError("resolve", err.Error()))
		}
		return data, diags

	}

	return nil, diags.Append(diag.Diagnostic{
		Severity: diag.SeverityError,
		Summary:  "Unsupported identifier",
		Detail:   fmt.Sprintf("Expected a valid identifier, got %s", id),
	})
}
