package ops

import (
	"fmt"
	"github.com/GateOfBabylon/enuma-elish-interpreter/types"
	"os"
	"regexp"
	"strings"
)

// ResolveExport parses the export pattern and sets the environment variable with the response.
func ResolveExport(exportStr string, response string) error {
	if exportStr == "" {
		return nil
	}

	exportPattern := regexp.MustCompile(`^\${{\s*([\w_]+)\s*}}$`)
	matches := exportPattern.FindStringSubmatch(exportStr)
	if len(matches) != 2 {
		return fmt.Errorf("invalid export pattern, expected format: `${{VAR_NAME}}`, got: %s", exportStr)
	}

	varName := matches[1]
	if varName == "" {
		return fmt.Errorf("export variable name is empty")
	}

	if err := os.Setenv(varName, response); err != nil {
		return fmt.Errorf("failed to set export variable %s: %w", varName, err)
	}
	return nil
}

// ResolveExportIntoOs handles exports in format 'VAR_NAME=VALUE'.
func ResolveExportIntoOs(exportStr string) error {
	parts := strings.SplitN(exportStr, "=", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid export format, expected 'VAR_NAME=VALUE', got: %s", exportStr)
	}

	left := strings.TrimSpace(parts[0])
	right := strings.TrimSpace(parts[1])

	exportPattern := regexp.MustCompile(`^\${{\s*([\w_]+)\s*}}$`)
	matches := exportPattern.FindStringSubmatch(left)
	if len(matches) != 2 {
		return fmt.Errorf("invalid export pattern, expected format: `${{VAR_NAME}}`, got: %s", left)
	}

	varName := matches[1]
	if varName == "" {
		return fmt.Errorf("export variable name is empty")
	}

	if err := os.Setenv(varName, right); err != nil {
		return fmt.Errorf("failed to set export variable %s: %w", varName, err)
	}

	return nil
}

func ExtractExportedValue(output string) (string, error) {
	exportPattern := regexp.MustCompile(`(?m)export\s+(\w+)\s*=\s*(.+)`)
	matches := exportPattern.FindStringSubmatch(output)

	if len(matches) != 3 {
		return "", fmt.Errorf("failed to extract export value from output")
	}

	varValue := strings.TrimSpace(matches[2])
	return varValue, nil
}

// ReplaceEnvsInTask replaces all environment variable placeholders in the task fields.
func ReplaceEnvsInTask(task *types.Task) {
	if task == nil {
		return
	}

	// Replace in HTTP task fields
	if task.HttpTaskFields != nil {
		task.HttpTaskFields.Method = resolveEnvVars(task.HttpTaskFields.Method)
		task.HttpTaskFields.Url = resolveEnvVars(task.HttpTaskFields.Url)
		task.HttpTaskFields.Body = resolveEnvVars(task.HttpTaskFields.Body)

		for key, value := range task.HttpTaskFields.Headers {
			task.HttpTaskFields.Headers[key] = resolveEnvVars(value)
		}
	}

	// Replace in Python task fields
	if task.PyTaskFields != nil {
		task.PyTaskFields.ScriptPath = resolveEnvVars(task.PyTaskFields.ScriptPath)
	}

	// Replace in ConditionStatements
	if task.ConditionStatements != nil {
		task.ConditionStatements.Condition = resolveEnvVars(task.ConditionStatements.Condition)

		// Replace in Picks
		if task.ConditionStatements.Pick != nil {
			if task.ConditionStatements.Pick.Else != nil {
				ReplaceEnvsInTask(task.ConditionStatements.Pick.Else)
			}
			for _, ifStmt := range task.ConditionStatements.Pick.IfStatement {
				ifStmt.Try = resolveEnvVars(ifStmt.Try)
				if ifStmt.Task != nil {
					ReplaceEnvsInTask(ifStmt.Task)
				}
			}
		}

		// Replace in Parallel tasks
		if task.ConditionStatements.Parallel != nil {
			for i := range *task.ConditionStatements.Parallel {
				ReplaceEnvsInTask(&(*task.ConditionStatements.Parallel)[i])
			}
		}
	}
}

// resolveEnvVars replaces all environment variable placeholders in the input string.
func resolveEnvVars(input string) string {
	re := regexp.MustCompile(`\${{\s*([^{}]+)\s*}}`)
	matches := re.FindAllStringSubmatch(input, -1)

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		envVar := strings.TrimSpace(match[1])
		value, exists := os.LookupEnv(envVar)
		if !exists {
			continue
		}
		input = strings.ReplaceAll(input, match[0], value)
	}
	return input
}
