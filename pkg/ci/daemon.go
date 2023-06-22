package ci

import (
	"context"
	"github.com/hashicorp/hcl/v2"
	"github.com/srevinsaju/togomak/v1/pkg/c"
)

type DaemonLifecycle struct {
	StopWhenComplete Blocks
}

type Lifecycle struct {
	StopWhenComplete hcl.Expression `hcl:"stop_when_complete,optional" json:"stop_when_complete"`
}

func (l *Lifecycle) Parse(ctx context.Context) (*DaemonLifecycle, hcl.Diagnostics) {

	pipe := ctx.Value(c.TogomakContextPipeline).(*Pipeline)
	daemonLifecycle := &DaemonLifecycle{}
	var diags hcl.Diagnostics

	if l == nil || l.StopWhenComplete == nil {
		return daemonLifecycle, diags
	}
	variables := l.StopWhenComplete.Variables()

	var runnableString []string
	for _, v := range variables {
		data, d := ResolveFromTraversal(v)
		diags = diags.Extend(d)
		if data == "" || diags.HasErrors() {
			continue
		}
		runnableString = append(runnableString, data)
	}
	var runnables Blocks
	for _, runnableId := range runnableString {
		runnable, diags := Resolve(ctx, pipe, runnableId)
		if diags.HasErrors() {
			return nil, diags
		}
		runnables = append(runnables, runnable)
	}
	daemonLifecycle.StopWhenComplete = runnables

	return daemonLifecycle, diags
}

func (l *Lifecycle) Variables() []hcl.Traversal {
	return nil
}
