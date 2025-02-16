package engine

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
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
		return fmt.Errorf("python executor type not implemented yet")
	default:
		return fmt.Errorf("executor type %s not implemented", executorType)
	}
}

// executeHTTPTask handles HTTP task execution.
func executeHTTPTask(task *types.Task) error {
	fmt.Printf("Executing HTTP task: %s\n", task.Name)
	replaceEnvsInTask(task)

	executeFunc := func() (string, error) {
		if task.HttpTaskFields != nil {
			return executeHTTPRequest(task.HttpTaskFields)
		}
		return "", executeDefaultTask(task)
	}

	if task.TimeStatements != nil && task.TimeStatements.Timeout > 0 {
		timeoutDuration := time.Duration(task.TimeStatements.Timeout) * time.Millisecond
		executeFunc = func() (string, error) {
			return executeWithTimeout(executeFunc, timeoutDuration)
		}
	}

	retries := 0
	if task.TimeStatements != nil {
		retries = task.TimeStatements.Retry
	}

	response, err := executeWithRetry(executeFunc, retries)
	if err != nil {
		return err
	}

	if task.TimeStatements != nil && task.TimeStatements.Delay > 0 {
		time.Sleep(time.Duration(task.TimeStatements.Delay) * time.Millisecond)
	}
	fmt.Println("Response before export: " + response)
	if task.Export != "" && response != "" {
		return resolveExport(task.Export, response)
	}

	return nil
}

// executeDefaultTask handles generic task execution, including parallel tasks.xq
func executeDefaultTask(task *types.Task) error {
	fmt.Printf("Executing task: %s\n", task.Name)

	if task.ConditionStatements.Parallel != nil {
		err := executeParallelTasks(task.ConditionStatements.Parallel)
		return err
	}
	return fmt.Errorf("no execution logic defined for task: %s", task.Name)
}

// executeParallelTasks executes multiple tasks concurrently.
func executeParallelTasks(parallelTasks *[]types.Task) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(*parallelTasks))

	for _, subTask := range *parallelTasks {
		wg.Add(1)
		go func(t types.Task) {
			defer wg.Done()
			result, err := executeHTTPRequest(t.HttpTaskFields)
			if err != nil {
				errChan <- err
			} else {
				errChan <- resolveExport(subTask.Export, result)
			}
		}(subTask)
	}

	wg.Wait()
	close(errChan)

	errorSlice := make([]string, 0)
	for err := range errChan {
		if err != nil {
			errorSlice = append(errorSlice, err.Error())
		}
	}

	if len(errorSlice) > 0 {
		return fmt.Errorf("parallel task execution encountered errors: %s", strings.Join(errorSlice, "; "))
	}

	return nil
}

// executeHTTPRequest performs the HTTP request based on the provided HttpTaskFields.
func executeHTTPRequest(httpFields *types.HttpTaskFields) (string, error) {
	var body io.Reader
	if httpFields.Body != "" {
		body = strings.NewReader(httpFields.Body)
	}

	req, err := http.NewRequest(httpFields.Method, httpFields.Url, body)
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}

	for key, value := range httpFields.Headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	defer resp.Body.Close()

	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	return string(responseData), nil
}

// /////////////////////
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
func replaceEnvsInTask(task *types.Task) {
	if task == nil {
		return
	}

	// Replace in HTTP task fields
	if task.HttpTaskFields != nil {
		task.HttpTaskFields.Method = resolveEnvVars(task.HttpTaskFields.Method)
		task.HttpTaskFields.Url = resolveEnvVars(task.HttpTaskFields.Url)
		task.HttpTaskFields.Body = resolveEnvVars(task.HttpTaskFields.Body)

		// Replace in headers
		for key, value := range task.HttpTaskFields.Headers {
			task.HttpTaskFields.Headers[key] = resolveEnvVars(value)
		}
	}

	// Replace in ConditionStatements
	if task.ConditionStatements != nil {
		task.ConditionStatements.Condition = resolveEnvVars(task.ConditionStatements.Condition)

		// Replace in Picks
		if task.ConditionStatements.Pick != nil {
			if task.ConditionStatements.Pick.Else != nil {
				replaceEnvsInTask(task.ConditionStatements.Pick.Else)
			}
			for _, ifStmt := range task.ConditionStatements.Pick.IfStatement {
				ifStmt.Try = resolveEnvVars(ifStmt.Try)
				if ifStmt.Task != nil {
					replaceEnvsInTask(ifStmt.Task)
				}
			}
		}

		// Replace in Parallel tasks
		if task.ConditionStatements.Parallel != nil {
			for i := range *task.ConditionStatements.Parallel {
				replaceEnvsInTask(&(*task.ConditionStatements.Parallel)[i])
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
			fmt.Printf("environment variable %s not found\n", envVar)
			continue
		}
		input = strings.ReplaceAll(input, match[0], value)
	}
	return input
}
