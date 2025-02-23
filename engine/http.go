package engine

import (
	"fmt"
	"github.com/GateOfBabylon/enuma-elish-interpreter/engine/ops"
	"github.com/GateOfBabylon/enuma-elish-interpreter/types"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// executeHTTPTask handles HTTP task execution.
func executeHTTPTask(task *types.Task) error {
	log.Printf("Executing HTTP task: %q\n", task.Name)
	ops.ReplaceEnvsInTask(task)

	executeFunc := func() (string, error) {
		if task.HttpTaskFields != nil {
			return executeHTTPRequest(task.HttpTaskFields)
		}
		return "", executeDefaultTask(task, &types.Universe{})
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

	if task.Export != "" && response != "" {
		log.Printf("Exporting response for Task: %q\n", task.Name)
		return ops.ResolveExport(task.Export, response)
	}

	log.Printf("Finished HTTP task: %q\n", task.Name)
	return nil
}

// executeHTTPRequest performs the HTTP request based on the provided HttpTaskFields.
func executeHTTPRequest(httpFields *types.HttpTaskFields) (string, error) {
	log.Printf("Creating request %s %s\n", httpFields.Method, httpFields.Url)

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
