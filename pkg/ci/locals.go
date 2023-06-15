package ci

import (
	"fmt"
	"github.com/srevinsaju/togomak/v1/pkg/diag"
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

func (l Locals) Expand() ([]*Local, diag.Diagnostics) {
	var locals []*Local
	var diags diag.Diagnostics
	attr, err := l.Body.JustAttributes()
	if err != nil {
		return nil, diags.Append(diag.Diagnostic{
			Severity: diag.SeverityError,
			Summary:  "Failed to decode locals",
			Detail:   err.Error(),
			Source:   l.Identifier(),
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

func (*Local) Identifier() string {
	return fmt.Sprintf("local.%s", &Local{})
}

func (*Local) Type() string {
	return LocalBlock
}

func (*Local) IsDaemon() bool {
	return false
}

func (*Local) Terminate() diag.Diagnostics {
	return nil
}

func (*Local) Kill() diag.Diagnostics {
	return nil
}

func (l LocalsGroup) Expand() ([]*Local, diag.Diagnostics) {
	var diags diag.Diagnostics
	var locals []*Local
	for _, lo := range l {
		ll, dd := lo.Expand()
		diags = diags.Extend(dd)
		locals = append(locals, ll...)
	}
	return locals, diags
}

func (l LocalGroup) ById(id string) (*Local, error) {
	for _, lo := range l {
		if lo.Key == id {
			return lo, nil
		}
	}
	return nil, fmt.Errorf("local variable with id %s not found", id)
}
