package engine

import (
	"fmt"
	"github.com/GateOfBabylon/enuma-elish-interpreter/types"
	"github.com/GateOfBabylon/enuma-elish-interpreter/validate"
)

func ExecuteTask(task *types.Task, executorType types.ExecutorType) error {
	// Validate the task
	if err := validate.ValidateTask(task, executorType); err != nil {
		return fmt.Errorf("task validation failed: %w", err)
	}

	fmt.Printf("Executing task: %s\n", task.Name)

	//// Example task execution logic (mocked for simplicity)
	//if task.Iterate != "" {
	//	fmt.Printf("Iterating over %s\n", task.Iterate)
	//}
	//
	//// Handle parallel tasks
	//if task.Parallel != nil {
	//	fmt.Println("Executing parallel tasks")
	//	for _, parallelTask := range *task.Parallel {
	//		if err := ExecuteTask(&parallelTask); err != nil {
	//			return fmt.Errorf("parallel task %s failed: %w", parallelTask.Name, err)
	//		}
	//	}
	//}
	//
	//// Simulate HTTP request execution (replace with actual implementation)
	//fmt.Printf("Making %s request to %s\n", task.Method, task.Url)
	return nil
}
