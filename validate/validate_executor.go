package validate

import (
	"errors"
	"fmt"
	"github.com/GateOfBabylon/enuma-elish-interpreter/types"
)

type Validator interface {
	Validate(exec *types.Executor) error
}

var ValidationFactory = make(map[types.ExecutorType]Validator)

func RegisterValidator(executorType types.ExecutorType, validator Validator) {
	ValidationFactory[executorType] = validator
}

func GetValidator(execType types.ExecutorType) (Validator, error) {
	validator, exists := ValidationFactory[execType]
	if !exists {
		return nil, errors.New(fmt.Sprintf("validator for executor of type %s does not exist", execType))
	}
	return validator, nil
}

type PythonExecutorValidator struct{}

func (pev *PythonExecutorValidator) Validate(exec *types.Executor) error {
	if exec.Universe != nil {
		return errors.New("universe property is not supported for Python executor")
	}
	return nil
}

type HttpExecutorValidator struct{}

func (hev *HttpExecutorValidator) Validate(exec *types.Executor) error {
	if exec.Universe == nil {
		return errors.New("universe property is mandatory for HTTP executor")
	}
	return nil
}

func init() {
	RegisterValidator(types.PYTHON, &PythonExecutorValidator{})
	RegisterValidator(types.HTTP, &HttpExecutorValidator{})
}

func ValidateExecutor(executor *types.Executor) error {
	validator, err := GetValidator(executor.Type)
	if err != nil {
		return err
	}
	return validator.Validate(executor)
}
