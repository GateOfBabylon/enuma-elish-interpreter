package engine

import (
	"fmt"
	"io"
	"log"
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
	log.Printf("Starting execution for Task: %q with ExecutorType: %s\n", task.Name, executorType)

	if err := validate.ValidateTask(task, executorType); err != nil {
		return fmt.Errorf("task validation failed: %w", err)
	}

	// Dispatch execution based on executor type
	switch executorType {
	case types.HTTP:
		log.Printf("Routing to HTTP task executor for Task: %q\n", task.Name)
		return executeHTTPTask(task)
	case types.PYTHON:
		return fmt.Errorf("python executor type not implemented yet")
	default:
		return fmt.Errorf("executor type %s not implemented", executorType)
	}
}

// executeHTTPTask handles HTTP task execution.
func executeHTTPTask(task *types.Task) error {
	log.Printf("Executing HTTP task: %q\n", task.Name)
	ops.ReplaceEnvsInTask(task)

	executeFunc := func() (string, error) {
		if task.HttpTaskFields != nil {
			log.Printf("Found HttpTaskFields for Task: %q\n", task.Name)
			return executeHTTPRequest(task.HttpTaskFields)
		}
		log.Printf("No HttpTaskFields. Falling back to default task logic for Task: %q\n", task.Name)
		return "", executeDefaultTask(task)
	}

	// Wrap with timeout if specified
	if task.TimeStatements != nil && task.TimeStatements.Timeout > 0 {
		timeoutDuration := time.Duration(task.TimeStatements.Timeout) * time.Millisecond
		log.Printf("Applying timeout of %d ms for Task: %q\n", task.TimeStatements.Timeout, task.Name)
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
	log.Printf("Starting HTTP request logic for Task: %q\n", task.Name)
	response, err := ops.ExecuteWithRetry(executeFunc, retries)
	if err != nil {
		return fmt.Errorf("failed after retries for Task: %q: %w", task.Name, err)
	}

	// Apply delay if specified
	if task.TimeStatements != nil && task.TimeStatements.Delay > 0 {
		log.Printf("Applying delay of %d ms for Task: %q\n", task.TimeStatements.Delay, task.Name)
		time.Sleep(time.Duration(task.TimeStatements.Delay) * time.Millisecond)
	}

	log.Printf("Task: %q received response: %q\n", task.Name, response)
	if task.Export != "" && response != "" {
		log.Printf("Exporting response for Task: %q\n", task.Name)
		return ops.ResolveExport(task.Export, response)
	}

	log.Printf("Finished HTTP task: %q with no export\n", task.Name)
	return nil
}

// executeDefaultTask handles generic task execution, including parallel tasks, condition/pick statements.
func executeDefaultTask(task *types.Task) error {
	log.Printf("Executing default task logic for Task: %q\n", task.Name)

	if task.ConditionStatements != nil {
		if task.ConditionStatements.Parallel != nil {
			log.Printf("Found parallel tasks for: %q\n", task.Name)
			return executeParallelTasks(task.ConditionStatements.Parallel)
		}
		if task.ConditionStatements.Pick != nil {
			log.Printf("Found pick statements for: %q\n", task.Name)
			return executePickStatement(task.ConditionStatements.Pick)
		}
	}

	// If there's no special logic, but we have an export directive
	if task.Export != "" {
		log.Printf("Exporting environment variable for: %q\n", task.Name)
		return ops.ResolveExportIntoOs(task.Export)
	}

	return fmt.Errorf("no execution logic defined for task: %q", task.Name)
}

// executeHTTPRequest performs the HTTP request based on the provided HttpTaskFields.
func executeHTTPRequest(httpFields *types.HttpTaskFields) (string, error) {
	log.Printf("Creating request %s %s\n", httpFields.Method, httpFields.Url)

	var body io.Reader
	if httpFields.Body != "" {
		log.Printf("With body: %s\n", httpFields.Body)
		body = strings.NewReader(httpFields.Body)
	}

	req, err := http.NewRequest(httpFields.Method, httpFields.Url, body)
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}

	for key, value := range httpFields.Headers {
		log.Printf("Setting header: %s = %s\n", key, value)
		req.Header.Set(key, value)
	}

	client := &http.Client{}
	log.Printf("Sending request...\n")
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	defer resp.Body.Close()

	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	log.Printf("Received status: %s\n", resp.Status)
	return string(responseData), nil
}

// executeParallelTasks executes multiple tasks concurrently.
func executeParallelTasks(parallelTasks *[]types.Task) error {
	log.Printf("Found %d parallel tasks\n", len(*parallelTasks))
	var wg sync.WaitGroup
	errChan := make(chan error, len(*parallelTasks))

	for _, subTask := range *parallelTasks {
		wg.Add(1)
		go func(t types.Task) {
			defer wg.Done()
			log.Printf("Running subTask: %q\n", t.Name)

			if t.HttpTaskFields != nil {
				result, err := executeHTTPRequest(t.HttpTaskFields)
				if err != nil {
					errChan <- fmt.Errorf("error in subTask: %q => %w", t.Name, err)
					return
				}
				err = ops.ResolveExport(t.Export, result)
				if err != nil {
					errChan <- fmt.Errorf("export error in subTask: %q => %w", t.Name, err)
				}
			} else {
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
		return fmt.Errorf("encountered errors: %s", strings.Join(errors, "; "))
	}

	log.Printf("all parallel tasks completed successfully\n")
	return nil
}

// executePickStatement evaluates the if-else conditions and executes the appropriate task.
func executePickStatement(pick *types.PickStatement) error {
	log.Printf("evaluating if-conditions...\n")
	for _, statement := range pick.IfStatement {
		if ops.CalculateCondition(statement.Try) {
			log.Printf("condition was true: %q\n", statement.Try)
			task := statement.Task
			return runTask(task)
		} else {
			log.Printf("condition was false: %q\n", statement.Try)
		}
	}

	// If no if-condition is met, run the else block if it exists
	if pick.Else != nil {
		log.Printf("No if-condition matched, running else block\n")
		return runTask(pick.Else)
	}

	log.Printf("No conditions matched and no else block specified\n")
	return nil
}

// runTask decides whether the task is HTTP or default
func runTask(task *types.Task) error {
	log.Printf("Checking type of subTask: %q\n", task.Name)
	if task.PyTaskFields != nil {
		return fmt.Errorf("python not implemented yet")
	} else if task.HttpTaskFields != nil {
		log.Printf("SubTask: %q is HTTP\n", task.Name)
		return executeHTTPTask(task)
	}
	log.Printf("SubTask: %q is default/pick\n", task.Name)
	return executeDefaultTask(task)
}
