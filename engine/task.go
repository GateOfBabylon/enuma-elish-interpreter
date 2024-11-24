package engine

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/GateOfBabylon/enuma-elish-interpreter/types"
	"github.com/GateOfBabylon/enuma-elish-interpreter/validate"
)

// ExecuteTask is the main entry point for executing a task based on the executor type.
func ExecuteTask(task *types.Task, executorType types.ExecutorType) error {
	// Validate the task
	if err := validate.ValidateTask(task, executorType); err != nil {
		return fmt.Errorf("task validation failed: %w", err)
	}

	// Dispatch execution based on executor type
	switch executorType {
	case types.HTTP:
		return executeHTTPTask(task)
	case types.PYTHON:
		return fmt.Errorf("Python executor type not implemented yet")
	default:
		return fmt.Errorf("executor type %s not implemented", executorType)
	}
}

// executeHTTPTask handles the execution of HTTP tasks.
func executeHTTPTask(task *types.Task) error {
	fmt.Printf("Executing HTTP task: %s\n", task.Name)

	// Replace environment variables in the task
	if err := replaceEnvsInTask(task); err != nil {
		return fmt.Errorf("failed to replace environment variables: %w", err)
	}

	// Prepare the execution function
	executeFunc := func() (string, error) {
		return executeHTTPRequest(task.HttpTaskFields)
	}

	// Wrap the execution with timeout if specified
	if task.TimeStatements != nil && task.TimeStatements.Timeout > 0 {
		timeoutDuration := time.Duration(task.TimeStatements.Timeout) * time.Millisecond
		executeFunc = func() (string, error) {
			return executeWithTimeout(executeFunc, timeoutDuration)
		}
	}

	// Wrap the execution with retries if specified
	var retries int
	if task.TimeStatements != nil {
		retries = task.TimeStatements.Retry
	}
	response, err := executeWithRetry(executeFunc, retries)
	if err != nil {
		return err
	}

	// Handle delay after execution if specified
	if task.TimeStatements != nil && task.TimeStatements.Delay > 0 {
		time.Sleep(time.Duration(task.TimeStatements.Delay) * time.Millisecond)
	}

	// Handle export if specified
	if task.Export != "" {
		if err := resolveExport(task.Export, response); err != nil {
			return err
		}
	}

	return nil
}

// executeHTTPRequest performs the HTTP request based on the provided HttpTaskFields.
func executeHTTPRequest(httpFields *types.HttpTaskFields) (string, error) {
	// Prepare the request body
	var body io.Reader
	if httpFields.Body != "" {
		body = strings.NewReader(httpFields.Body)
	}

	// Create a new HTTP request
	req, err := http.NewRequest(httpFields.Method, httpFields.Url, body)
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set the headers
	for key, value := range httpFields.Headers {
		req.Header.Set(key, value)
	}

	// Execute the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	return string(responseData), nil
}

// executeWithTimeout runs a function with a timeout.
func executeWithTimeout(executeFunc func() (string, error), timeout time.Duration) (string, error) {
	resultChan := make(chan string)
	errorChan := make(chan error)

	go func() {
		response, err := executeFunc()
		if err != nil {
			errorChan <- err
			return
		}
		resultChan <- response
	}()

	select {
	case response := <-resultChan:
		return response, nil
	case err := <-errorChan:
		return "", err
	case <-time.After(timeout):
		return "", fmt.Errorf("task execution timed out after %s", timeout)
	}
}

// executeWithRetry retries a function execution based on the specified retry count.
func executeWithRetry(executeFunc func() (string, error), retries int) (string, error) {
	var response string
	var err error

	for attempt := 0; attempt <= retries; attempt++ {
		response, err = executeFunc()
		if err == nil {
			return response, nil
		}

		fmt.Printf("Attempt %d failed: %v\n", attempt+1, err)
		if attempt < retries {
			time.Sleep(1 * time.Second) // Default delay between retries
			fmt.Printf("Retrying...\n")
		}
	}

	return "", fmt.Errorf("all %d attempts failed: %w", retries+1, err)
}

// resolveExport parses the export pattern and sets the environment variable with the response.
func resolveExport(exportStr string, response string) error {
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

	fmt.Printf("Exported environment variable: %s = %s\n", varName, response)
	return nil
}

// replaceEnvsInTask replaces all environment variable placeholders in the task fields.
func replaceEnvsInTask(task *types.Task) error {
	if task == nil {
		return nil
	}

	var err error

	// Replace in HTTP task fields
	if task.HttpTaskFields != nil {
		if task.HttpTaskFields.Method, err = resolveEnvVars(task.HttpTaskFields.Method); err != nil {
			return err
		}
		if task.HttpTaskFields.Url, err = resolveEnvVars(task.HttpTaskFields.Url); err != nil {
			return err
		}
		if task.HttpTaskFields.Body, err = resolveEnvVars(task.HttpTaskFields.Body); err != nil {
			return err
		}

		// Replace in headers
		for key, value := range task.HttpTaskFields.Headers {
			if task.HttpTaskFields.Headers[key], err = resolveEnvVars(value); err != nil {
				return err
			}
		}
	}

	// Replace in ConditionStatements
	if task.ConditionStatements != nil {
		if task.ConditionStatements.Iterate, err = resolveEnvVars(task.ConditionStatements.Iterate); err != nil {
			return err
		}
		if task.ConditionStatements.Condition, err = resolveEnvVars(task.ConditionStatements.Condition); err != nil {
			return err
		}

		// Replace in Picks
		if task.ConditionStatements.Pick != nil {
			if task.ConditionStatements.Pick.Else != nil {
				if err := replaceEnvsInTask(task.ConditionStatements.Pick.Else); err != nil {
					return err
				}
			}
			for _, ifStmt := range task.ConditionStatements.Pick.IfStatement {
				if ifStmt.Try, err = resolveEnvVars(ifStmt.Try); err != nil {
					return err
				}
				if ifStmt.Task != nil {
					if err := replaceEnvsInTask(ifStmt.Task); err != nil {
						return err
					}
				}
			}
		}

		// Replace in Parallel tasks
		if task.ConditionStatements.Parallel != nil {
			for i := range *task.ConditionStatements.Parallel {
				if err := replaceEnvsInTask(&(*task.ConditionStatements.Parallel)[i]); err != nil {
					return err
				}
			}
		}
	}

	// Replace in other fields
	if task.Name, err = resolveEnvVars(task.Name); err != nil {
		return err
	}
	if task.Export, err = resolveEnvVars(task.Export); err != nil {
		return err
	}

	return nil
}

// resolveEnvVars replaces all environment variable placeholders in the input string.
func resolveEnvVars(input string) (string, error) {
	re := regexp.MustCompile(`\${{([^{}]+)}}`)
	matches := re.FindAllStringSubmatch(input, -1)

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		envVar := strings.TrimSpace(match[1])
		value, exists := os.LookupEnv(envVar)
		if !exists {
			return "", fmt.Errorf("environment variable %s not found", envVar)
		}
		input = strings.Replace(input, match[0], value, -1)
	}
	return input, nil
}
