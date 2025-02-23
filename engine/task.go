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
		log.Printf("Routing to HTTP task executor for Task: %q\n", task.Name)
		return executeHTTPTask(task)
	case types.PYTHON:
		log.Printf("Routing to Python task executor for Task: %q\n", task.Name)
		return executePythonTask(task, executor.Universe)
	default:
		return fmt.Errorf("executor type %s not implemented", executor.Type)
	}
}

// executeDefaultTask handles generic task execution, including parallel tasks, condition/pick statements.
func executeDefaultTask(task *types.Task, universe *types.Universe) error {
	log.Printf("Executing default task logic for Task: %q\n", task.Name)

	if task.ConditionStatements != nil {
		if task.ConditionStatements.Parallel != nil {
			log.Printf("Found parallel tasks for: %q\n", task.Name)
			return executeParallelTasks(task.ConditionStatements.Parallel, universe)
		}
		if task.ConditionStatements.Pick != nil {
			log.Printf("Found pick statements for: %q\n", task.Name)
			return executePickStatement(task.ConditionStatements.Pick, universe)
		}
	}

	// If there's no special logic, but we have an export directive
	if task.Export != "" {
		log.Printf("Exporting environment variable for: %q\n", task.Name)
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

	log.Printf("all parallel tasks completed successfully\n")
	return nil
}

// executePickStatement evaluates the if-else conditions and executes the appropriate task.
func executePickStatement(pick *types.PickStatement, universe *types.Universe) error {
	log.Printf("evaluating if-conditions...\n")
	for _, statement := range pick.IfStatement {
		if ops.CalculateCondition(statement.Try) {
			log.Printf("condition was true: %q\n", statement.Try)
			task := statement.Task
			return runTask(task, universe)
		} else {
			log.Printf("condition was false: %q\n", statement.Try)
		}
	}

	// If no if-condition is met, run the else block if it exists
	if pick.Else != nil {
		log.Printf("No if-condition matched, running else block\n")
		return runTask(pick.Else, universe)
	}

	log.Printf("No conditions matched and no else block specified\n")
	return nil
}

// runTask decides whether the task is HTTP or default
func runTask(task *types.Task, universe *types.Universe) error {
	log.Printf("Checking type of subTask: %q\n", task.Name)
	if task.PyTaskFields != nil {
		log.Printf("SubTask: %q is Python\n", task.Name)
		return executePythonTask(task, universe)
	} else if task.HttpTaskFields != nil {
		log.Printf("SubTask: %q is HTTP\n", task.Name)
		return executeHTTPTask(task)
	}
	log.Printf("SubTask: %q is default/pick\n", task.Name)
	return executeDefaultTask(task, universe)
}
