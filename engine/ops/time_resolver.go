package ops

import (
	"fmt"
	"time"
)

// ExecuteWithTimeout runs a function with a timeout.
func ExecuteWithTimeout(executeFunc func() (string, error), timeout time.Duration) (string, error) {
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

// ExecuteWithRetry retries a function execution based on the specified retry count.
func ExecuteWithRetry(executeFunc func() (string, error), retries int) (string, error) {
	var response string
	var err error

	for attempt := 0; attempt <= retries; attempt++ {
		response, err = executeFunc()
		if err == nil {
			return response, nil
		}
	}
	return "", fmt.Errorf("all %d attempts failed: %w", retries+1, err)
}
