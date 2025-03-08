package engine

import (
	"fmt"
	"github.com/GateOfBabylon/enuma-elish-interpreter/engine/ops"
	. "github.com/GateOfBabylon/enuma-elish-interpreter/logger"
	"github.com/GateOfBabylon/enuma-elish-interpreter/types"
	"io"
	"net/http"
	"strings"
	"time"
)

// executeHTTPTask handles HTTP task execution.
func executeHTTPTask(task *types.Task) error {
	logger := GetLogger()
	logger.Log("Executing HTTP task: %q", task.Name)
	ops.ReplaceEnvsInTask(task)

	executeFunc := func() (string, error) {
		if task.HttpTaskFields != nil {
			return executeHTTPRequest(task.HttpTaskFields)
		}
		return "", executeDefaultTask(task, &types.Universe{})
	}

	if task.TimeStatements != nil && task.TimeStatements.Timeout > 0 {
		timeoutDuration := time.Duration(task.TimeStatements.Timeout) * time.Millisecond
		logger.Log("Applying timeout of %d ms for Task: %q", task.TimeStatements.Timeout, task.Name)
		executeFunc = func() (string, error) {
			return ops.ExecuteWithTimeout(executeFunc, timeoutDuration)
		}
	}

	retries := 0
	if task.TimeStatements != nil {
		retries = task.TimeStatements.Retry
	}

	response, err := ops.ExecuteWithRetry(executeFunc, retries)
	if err != nil {
		return fmt.Errorf("failed after retries for Task: %q: %w", task.Name, err)
	}

	if task.TimeStatements != nil && task.TimeStatements.Delay > 0 {
		logger.Log("Applying delay of %d ms for Task: %q", task.TimeStatements.Delay, task.Name)
		time.Sleep(time.Duration(task.TimeStatements.Delay) * time.Millisecond)
	}

	if response != "" {
		logger.Log("Exporting response: %s", response)
		if task.Export != "" {
			return ops.ResolveExport(task.Export, response)
		}
	}

	logger.Log("Finished HTTP request: %q with SUCCESS.", task.Name)
	return nil
}

// executeHTTPRequest performs the HTTP request based on the provided HttpTaskFields.
func executeHTTPRequest(httpFields *types.HttpTaskFields) (string, error) {
	logger := GetLogger()
	logger.Log("Starting %s request for %s", httpFields.Method, httpFields.Url)

	var body io.Reader
	if httpFields.Body != "" {
		body = strings.NewReader(httpFields.Body)
	}

	req, err := http.NewRequest(httpFields.Method, httpFields.Url, body)
	if err != nil {
		logger.Log("Failed to create HTTP request: %w", err)
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}

	for key, value := range httpFields.Headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{}
	logger.Log("Sending request...")
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute HTTP request: %v", err)
	}
	defer resp.Body.Close()

	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	logger.Log("Request completed with status %d", resp.StatusCode)
	return string(responseData), nil
}
