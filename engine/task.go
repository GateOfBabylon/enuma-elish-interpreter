package engine

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/GateOfBabylon/enuma-elish-interpreter/engine/ops"
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
		// Future Python logic here
		return fmt.Errorf("python executor type not implemented yet")
	default:
		return fmt.Errorf("executor type %s not implemented", executorType)
	}
}

// executeHTTPTask handles HTTP task execution.
func executeHTTPTask(task *types.Task) error {
	fmt.Printf("Executing HTTP task: %s\n", task.Name)
	ops.ReplaceEnvsInTask(task)

	executeFunc := func() (string, error) {
		if task.HttpTaskFields != nil {
			return executeHTTPRequest(task.HttpTaskFields)
		}
		// Fallback if no HttpTaskFields -> go to default
		return "", executeDefaultTask(task)
	}

	// Wrap with timeout if specified
	if task.TimeStatements != nil && task.TimeStatements.Timeout > 0 {
		timeoutDuration := time.Duration(task.TimeStatements.Timeout) * time.Millisecond
		executeFunc = func() (string, error) {
			return ops.ExecuteWithTimeout(executeFunc, timeoutDuration)
		}
	}

	// Apply retry logic
	retries := 0
	if task.TimeStatements != nil {
		retries = task.TimeStatements.Retry
	}

	// Execute the function with optional retry
	response, err := ops.ExecuteWithRetry(executeFunc, retries)
	if err != nil {
		return err
	}

	// Apply delay if specified
	if task.TimeStatements != nil && task.TimeStatements.Delay > 0 {
		time.Sleep(time.Duration(task.TimeStatements.Delay) * time.Millisecond)
	}

	if task.Export != "" && response != "" {
		// Export the response to environment
		return ops.ResolveExport(task.Export, response)
	}

	return nil
}

// executeDefaultTask handles generic task execution, including parallel tasks, condition/pick statements.
func executeDefaultTask(task *types.Task) error {
	if task.ConditionStatements != nil {
		if task.ConditionStatements.Parallel != nil {
			return executeParallelTasks(task.ConditionStatements.Parallel)
		}
		if task.ConditionStatements.Pick != nil {
			return executePickStatement(task.ConditionStatements.Pick)
		}
	}

	// If there's no special logic, but we have an export directive
	if task.Export != "" {
		return ops.ResolveExportIntoOs(task.Export)
	}

	return fmt.Errorf("no execution logic defined for task: %s", task.Name)
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

// executeParallelTasks executes multiple tasks concurrently.
func executeParallelTasks(parallelTasks *[]types.Task) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(*parallelTasks))

	for _, subTask := range *parallelTasks {
		wg.Add(1)
		go func(t types.Task) {
			defer wg.Done()

			// Because it's a single HTTP sub-task or a default task
			// we can check if it's an HTTP or not
			if t.HttpTaskFields != nil {
				result, err := executeHTTPRequest(t.HttpTaskFields)
				if err != nil {
					errChan <- err
					return
				}
				// If subTask wants to export the result:
				err = ops.ResolveExport(t.Export, result)
				errChan <- err
			} else {
				// If it's a default or pick, etc.
				errChan <- executeDefaultTask(&t)
			}
		}(subTask)
	}

	wg.Wait()
	close(errChan)

	var errors []string
	for err := range errChan {
		if err != nil {
			errors = append(errors, err.Error())
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("parallel task execution encountered errors: %s", strings.Join(errors, "; "))
	}

	return nil
}

// executePickStatement evaluates the if-else conditions and executes the appropriate task.
func executePickStatement(pick *types.PickStatement) error {
	for _, statement := range pick.IfStatement {
		if ops.CalculateCondition(statement.Try) {
			fmt.Println("Condition was true: " + statement.Try)
			task := statement.Task
			return runTask(task)
		}
	}

	// If no if-condition is met, run the else block if exists
	if pick.Else != nil {
		fmt.Println("Running the else option")
		return runTask(pick.Else)
	}
	return nil
}

// runTask decides whether the task is HTTP or default
func runTask(task *types.Task) error {
	if task.PyTaskFields != nil {
		return fmt.Errorf("python not implemented yet")
	} else if task.HttpTaskFields != nil {
		return executeHTTPTask(task)
	}
	return executeDefaultTask(task)
}
