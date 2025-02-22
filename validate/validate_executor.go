package validate

import (
	"errors"
	"github.com/GateOfBabylon/enuma-elish-interpreter/types"
)

type PythonExecutorValidator struct{}

func (pev *PythonExecutorValidator) Validate(exec *types.Executor) error {
	if err := generalExecutorValidation(exec); err != nil {
		return err
	}
	if exec.Universe == nil {
		return errors.New("universe property is mandatory for Python executor")
	}
	if !(exec.Type == types.PYTHON) {
		return errors.New("executor type must be python")
	}
	return nil
}

type HttpExecutorValidator struct{}

func (hev *HttpExecutorValidator) Validate(exec *types.Executor) error {
	if err := generalExecutorValidation(exec); err != nil {
		return err
	}
	if exec.Universe != nil {
		return errors.New("universe property is not supported for HTTP executor")
	}
	if !(exec.Type == types.HTTP) {
		return errors.New("executor type must be http")
	}
	return nil
}

func generalExecutorValidation(exec *types.Executor) error {
	if exec.Name == "" {
		return errors.New("executor name is mandatory")
	}
	if !IsValidName(exec.Name) {
		return errors.New("executor name contains invalid characters")
	}
	if len(exec.Tasks) < 1 {
		return errors.New("executor has no tasks")
	}

	return nil
}

func init() {
	RegisterExecutorValidator(types.PYTHON, &PythonExecutorValidator{})
	RegisterExecutorValidator(types.HTTP, &HttpExecutorValidator{})
}

func Executor(executor *types.Executor) error {
	validator, err := GetExecutorValidator(executor.Type)
	if err != nil {
		return err
	}
	return validator.Validate(executor)
}
