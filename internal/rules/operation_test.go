package rules

import (
	"fmt"
	"testing"
)

func TestOperationUnmarshal(t *testing.T) {
	tests := []struct {
		input     string
		expected  Operation
		mustError bool
	}{
		{
			input: "+foo",
			expected: Operation{
				op:       OperationTypeAdd,
				runnable: "foo",
			},
		},
		{
			input: "^foo",
			expected: Operation{
				op:       OperationTypeSub,
				runnable: "foo",
			},
		},
		{
			input: "foo",
			expected: Operation{
				op:       OperationTypeAnd,
				runnable: "foo",
			},
		},
		{
			input:     "+",
			expected:  Operation{},
			mustError: true,
		},
		{
			input:     "^",
			expected:  Operation{},
			mustError: true,
		},
		{
			input:     "",
			expected:  Operation{},
			mustError: true,
		},
		{
			input:     "foo+",
			expected:  Operation{},
			mustError: true,
		},
		{
			input:     "foo^",
			expected:  Operation{},
			mustError: true,
		},
		{
			input:     "foo+bar",
			expected:  Operation{},
			mustError: true,
		},
		{
			input:     "foo^bar",
			expected:  Operation{},
			mustError: true,
		},
		{
			input:     "foo+bar^baz",
			expected:  Operation{},
			mustError: true,
		},
		{
			input: "+stage.bar",
			expected: Operation{
				op:       OperationTypeAdd,
				runnable: "stage.bar",
			},
		},
	}

	for i, test := range tests {
		fmt.Println(i, test.input)
		op, d := OperationUnmarshal(test.input)
		if test.mustError && !d.HasErrors() {
			t.Errorf("expected error for '%s'", test.input)
			continue
		}
		if !test.mustError && d.HasErrors() {
			t.Errorf("unexpected error for '%s'", test.input)
			continue
		}
		if test.mustError && d.HasErrors() {
			continue
		}
		if op.op != test.expected.op {
			t.Errorf("expected op %s, got %s for '%s'", test.expected.op, op.op, test.input)
			continue
		}
		if op.runnable != test.expected.runnable {
			t.Errorf("expected runnable %s, got %s for '%s'", test.expected.runnable, op.runnable, test.input)
			continue
		}
	}
}
