package engine

import (
	"fmt"
	"github.com/GateOfBabylon/enuma-elish-interpreter/convert"
	"github.com/GateOfBabylon/enuma-elish-interpreter/types"
	"github.com/GateOfBabylon/enuma-elish-interpreter/validate"
	"os"
)

func ExecuteExecutor(executor *types.Executor) error {
	if err := validate.Executor(executor); err != nil {
		return fmt.Errorf("executor validation failed: %w", err)
	}

	fmt.Printf("Executing executor: %s\n", executor.Name)
	if err := ExtractEnvs(executor); err != nil {
		return err
	}

	for _, task := range executor.Tasks {
		if err := ExecuteTask(&task, executor.Type); err != nil {
			return fmt.Errorf("task %q execution failed: %w", task.Name, err)
		}
	}
	return nil
}

func ExtractEnvs(executor *types.Executor) error {
	for key, value := range executor.Env {
		strValue := convert.InterfaceToString(value)
		if err := os.Setenv(key, strValue); err != nil {
			return fmt.Errorf("setting environment variable %q failed: %w", key, err)
		}
	}
	return nil
}
