package engine

import (
	"fmt"
	"log"
	"os"

	"github.com/GateOfBabylon/enuma-elish-interpreter/convert"
	"github.com/GateOfBabylon/enuma-elish-interpreter/types"
	"github.com/GateOfBabylon/enuma-elish-interpreter/validate"
)

func ExecuteExecutor(executor *types.Executor) error {
	if err := validate.Executor(executor); err != nil {
		return fmt.Errorf("executor validation failed: %w", err)
	}

	log.Printf("Executing executor: %s, type: %s\n", executor.Name, executor.Type)
	if err := extractEnvs(executor); err != nil {
		return err
	}

	for _, task := range executor.Tasks {
		if err := ExecuteTask(&task, executor); err != nil {
			return fmt.Errorf("task %q execution failed: %w", task.Name, err)
		}
	}
	return nil
}

func extractEnvs(executor *types.Executor) error {
	for key, value := range executor.Env {
		strValue := convert.InterfaceToString(value)
		if err := os.Setenv(key, strValue); err != nil {
			return fmt.Errorf("setting environment variable %q failed: %w", key, err)
		}
	}
	return nil
}
