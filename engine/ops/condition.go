package ops

import (
	"strings"
)

// operatorFunctions maps a string operator to its evaluation logic.
var operatorFunctions = map[string]func(string, string) bool{
	"==": func(lhs, rhs string) bool { return lhs == rhs },
	"!=": func(lhs, rhs string) bool { return lhs != rhs },
}

func CalculateCondition(condition string) bool {
	condition = strings.TrimSpace(condition)

	var chosenOp string
	for op := range operatorFunctions {
		if strings.Contains(condition, op) {
			chosenOp = op
			break
		}
	}

	if chosenOp == "" {
		return false
	}

	parts := strings.Split(condition, chosenOp)
	if len(parts) != 2 {
		return false
	}

	left := resolveEnvVars(strings.TrimSpace(parts[0]))
	right := resolveEnvVars(strings.TrimSpace(parts[1]))

	right = strings.Trim(right, "'\"")

	return operatorFunctions[chosenOp](left, right)
}
