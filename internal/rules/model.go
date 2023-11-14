package rules

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/srevinsaju/togomak/v1/internal/blocks"
	"regexp"
	"strings"
)

type OperationType int

const (
	OperationTypeNone OperationType = iota
	OperationTypeAdd
	OperationTypeSub
	OperationTypeAnd
)

var (
	operationAddMatcher = regexp.MustCompile(`^\+([a-zA-Z0-9.\-:_]+)$`)
	operationSubMatcher = regexp.MustCompile(`^\^([a-zA-Z0-9.\-:_]+)$`)

	operationAndMatcher = regexp.MustCompile(`^([a-zA-Z0-9.\-:_]+)$`)
)

var OperationTypes = []OperationType{
	OperationTypeAdd,
	OperationTypeSub,
	OperationTypeAnd,
}

func (op OperationType) String() string {
	switch op {
	case OperationTypeAdd:
		return "+"
	case OperationTypeSub:
		return "^"
	case OperationTypeAnd:
		return ""
	}

	return ""
}

func (op OperationType) Matcher() *regexp.Regexp {
	switch op {
	case OperationTypeAdd:
		return operationAddMatcher
	case OperationTypeSub:
		return operationSubMatcher
	case OperationTypeAnd:
		return operationAndMatcher
	}
	panic(fmt.Sprintf("invalid operation type: %d", op))
}

type Operation struct {
	op       OperationType
	runnable string
}

type Operations []*Operation

func NewOperation(op OperationType, runnable string) *Operation {
	return &Operation{
		op:       op,
		runnable: runnable,
	}
}

func (ops Operations) Children(runnableId string) Operations {
	childOps := make(Operations, 0)
	for _, op := range ops {
		genericParam := true
		for _, block := range []string{blocks.StageBlock, blocks.ModuleBlock, blocks.MacroBlock} {
			if strings.HasPrefix(op.runnable, fmt.Sprintf("%s.", block)) {
				genericParam = false
			}
		}
		if genericParam {
			childOps = append(childOps, op)
		}
	}
	return childOps
}

func (op *Operation) String() string {
	return fmt.Sprintf("%s%s", op.op.String(), op.runnable)
}

func (ops Operations) Marshall() []string {
	var s []string
	for _, op := range ops {
		s = append(s, op.String())
	}
	return s
}

func (op *Operation) RunnableId() string {
	return op.runnable
}

func (op *Operation) Operation() OperationType {
	return op.op
}

func OperationUnmarshal(arg string) (*Operation, hcl.Diagnostics) {
	op := OperationTypeNone
	var diags hcl.Diagnostics

	for _, opType := range OperationTypes {
		matcher := opType.Matcher()
		if matcher.MatchString(arg) {
			op = opType
			break
		}
	}
	if op == OperationTypeNone {
		return nil, diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "invalid operation",
			Detail:   fmt.Sprintf("invalid operation, no operation found in %s.", arg),
		})
	}
	item := op.Matcher().FindStringSubmatch(arg)
	if item[1] == "" {
		return nil, diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "invalid operation",
			Detail:   fmt.Sprintf("invalid operation, no stage, module or macro followed operation '%s', found in %s.", op.String(), arg),
		})
	}
	return NewOperation(op, item[1]), nil
}

func Unmarshal(args []string) (ops Operations, diags hcl.Diagnostics) {
	// dslTokens := make([]string, 0)
	ops = make(Operations, len(args))
	var d hcl.Diagnostics
	for i, arg := range args {
		ops[i], d = OperationUnmarshal(arg)
		diags = diags.Extend(d)
	}
	return ops, diags
}
