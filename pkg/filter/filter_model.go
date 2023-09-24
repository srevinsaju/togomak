package filter

import "strings"

type Operation int64

const (
	OperationInvalid Operation = -1

	OperationAdd Operation = iota
	OperationSubtract

	OperationRun
)

func (op Operation) String() string {
	switch op {
	case OperationAdd:
		return "+"
	case OperationSubtract:
		return "^"
	case OperationRun:
		return ""
	}
	return ""
}

type Rule struct {
	// Type of block
	Type string

	// Operation is the operation that needs to be applied on the block
	Operation Operation

	// ID is the identifier of the block
	ID string
}

func Unmarshal(arg string) Rule {
	// Split the input string into parts using whitespace as the separator
	parts := strings.Fields(arg)

	// Initialize a Rule with default values
	rule := Rule{
		Type:      "",           // Default Type
		Operation: OperationRun, // Default Operation
		ID:        "",           // Default ID
	}

	// If there are no parts, return the default rule
	if len(parts) == 0 {
		return rule
	}

	// Extract the Operation based on the first part of the input string
	switch parts[0] {
	case "+":
		rule.Operation = OperationAdd
	case "^":
		rule.Operation = OperationSubtract
	}

	// If there is only one part (the Operation), return the rule with the Operation set
	if len(parts) == 1 {
		return rule
	}

	// If there are two parts, set the Type to the second part
	rule.Type = parts[1]

	// If there is a third part, set the ID to it
	if len(parts) >= 3 {
		rule.ID = parts[2]
	}

	return rule
}
