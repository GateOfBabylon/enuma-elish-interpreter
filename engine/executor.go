package engine

import (
	"fmt"
	"github.com/GateOfBabylon/enuma-elish-interpreter/types"
	"github.com/GateOfBabylon/enuma-elish-interpreter/validate"
)

func ExecuteExecutor(executor *types.Executor) error {
	if err := validate.Executor(executor); err != nil {
		return fmt.Errorf("executor validation failed: %w", err)
	}

	fmt.Printf("Executing executor: %s\n", executor.Name)
	for _, task := range executor.Tasks {
		if err := ExecuteTask(&task, executor.Type); err != nil {
			return fmt.Errorf("task %s execution failed: %w", task.Name, err)
		}
	}
	return nil
}
