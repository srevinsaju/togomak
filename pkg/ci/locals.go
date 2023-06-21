package ci

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
)

const LocalsBlock = "locals"
const LocalBlock = "local"

func (l Locals) Description() string {
	return ""
}

func (l Locals) Identifier() string {
	return fmt.Sprintf("locals.%p", &l)
}

func (l Locals) Type() string {
	return LocalsBlock
}

func (l Locals) Expand() ([]*Local, hcl.Diagnostics) {
	var locals []*Local
	var diags hcl.Diagnostics
	attr, err := l.Body.JustAttributes()
	if err != nil {
		return nil, diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Failed to decode locals",
			Detail:   err.Error(),
			Subject:  l.Body.MissingItemRange().Ptr(),
		})
	}

	for attr, expr := range attr {
		locals = append(locals, &Local{
			Key:   attr,
			Value: expr.Expr,
		})
	}
	return locals, diags
}

func (*Local) Description() string {
	return ""
}

func (l *Local) Identifier() string {
	return fmt.Sprintf("local.%s", l.Key)
}

func (*Local) Type() string {
	return LocalBlock
}

func (*Local) IsDaemon() bool {
	return false
}

func (*Local) Terminate() hcl.Diagnostics {
	return nil
}

func (*Local) Kill() hcl.Diagnostics {
	return nil
}

func (*Local) Set(k any, v any) {
}

func (*Local) Get(k any) any {
	return nil
}

func (l LocalsGroup) Expand() ([]*Local, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	var locals []*Local
	for _, lo := range l {
		ll, dd := lo.Expand()
		diags = diags.Extend(dd)
		locals = append(locals, ll...)
	}
	return locals, diags
}

func (l LocalGroup) ById(id string) (*Local, hcl.Diagnostics) {
	for _, lo := range l {
		if lo.Key == id {
			return lo, nil
		}
	}
	return nil, hcl.Diagnostics{
		{
			Severity: hcl.DiagError,
			Summary:  "Local not found",
			Detail:   "Local with id " + id + " not found",
		},
	}
}
