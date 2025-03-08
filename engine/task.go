package engine

import (
	"fmt"
	"github.com/GateOfBabylon/enuma-elish-interpreter/engine/ops"
	"github.com/GateOfBabylon/enuma-elish-interpreter/types"
	"github.com/GateOfBabylon/enuma-elish-interpreter/validate"
	"log"
	"strings"
	"sync"
)

func ExecuteTask(task *types.Task, executor *types.Executor) error {
	log.Printf("Starting execution for Task: %q with ExecutorType: %s\n", task.Name, executor.Type)

	// Validate the task
	if err := validate.Task(task, executor.Type); err != nil {
		return fmt.Errorf("task validation failed: %w", err)
	}

	switch executor.Type {
	case types.HTTP:
		return executeHTTPTask(task)
	case types.PYTHON:
		return executePythonTask(task, executor.Universe)
	default:
		return fmt.Errorf("executor type %s not implemented", executor.Type)
	}
}

// executeDefaultTask handles generic task execution, including parallel tasks, condition/pick statements.
func executeDefaultTask(task *types.Task, universe *types.Universe) error {
	if task.ConditionStatements != nil {
		if task.ConditionStatements.Parallel != nil {
			return executeParallelTasks(task.ConditionStatements.Parallel, universe)
		}
		if task.ConditionStatements.Pick != nil {
			return executePickStatement(task.ConditionStatements.Pick, universe)
		}
	}

	// If there's no special logic, but we have an export directive
	if task.Export != "" {
		return ops.ResolveExportIntoOs(task.Export)
	}

	return fmt.Errorf("no execution logic defined for task: %q", task.Name)
}

// executeParallelTasks executes multiple tasks concurrently.
func executeParallelTasks(parallelTasks *[]types.Task, universe *types.Universe) error {
	log.Printf("Found %d parallel tasks\n", len(*parallelTasks))
	var wg sync.WaitGroup
	errChan := make(chan error, len(*parallelTasks))

	for _, subTask := range *parallelTasks {
		wg.Add(1)
		go func(t types.Task) {
			defer wg.Done()
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
			} else if t.PyTaskFields != nil {
				result, err := executePythonScripts(t.PyTaskFields, universe, t.TimeStatements)
				if err != nil {
					errChan <- fmt.Errorf("error in subTask: %q => %w", t.Name, err)
					return
				}
				err = ops.ResolveExport(t.Export, result)
				if err != nil {
					errChan <- fmt.Errorf("export error in subTask: %q => %w", t.Name, err)
				}
			} else {
				errChan <- executeDefaultTask(&t, universe)
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
	return nil
}

// executePickStatement evaluates the if-else conditions and executes the appropriate task.
func executePickStatement(pick *types.PickStatement, universe *types.Universe) error {
	for _, statement := range pick.IfStatement {
		if ops.CalculateCondition(statement.Try) {
			task := statement.Task
			return runTask(task, universe)
		}
	}

	// If no if-condition is met, run the else block if it exists
	if pick.Else != nil {
		return runTask(pick.Else, universe)
	}

	return nil
}

// runTask decides whether the task is HTTP or default
func runTask(task *types.Task, universe *types.Universe) error {
	if task.PyTaskFields != nil {
		return executePythonTask(task, universe)
	} else if task.HttpTaskFields != nil {
		return executeHTTPTask(task)
	}
	return executeDefaultTask(task, universe)
}
