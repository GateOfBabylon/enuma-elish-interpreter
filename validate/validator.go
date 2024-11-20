package validate

import (
	"errors"
	"fmt"
	"github.com/GateOfBabylon/enuma-elish-interpreter/types"
)

// ExecutorValidator factory pattern
type ExecutorValidator interface {
	Validate(exec *types.Executor) error
}

var executorValidationFactory = make(map[types.ExecutorType]ExecutorValidator)

func RegisterExecutorValidator(executorType types.ExecutorType, validator ExecutorValidator) {
	executorValidationFactory[executorType] = validator
}

func GetExecutorValidator(execType types.ExecutorType) (ExecutorValidator, error) {
	validator, exists := executorValidationFactory[execType]
	if !exists {
		return nil, errors.New(fmt.Sprintf("validator for executor of type %s does not exist", execType))
	}
	return validator, nil
}

// TaskValidator factory pattern
type TaskValidator interface {
	Validate(task *types.Task) error
}

var taskValidationFactory = make(map[types.ExecutorType]TaskValidator)

func RegisterTaskValidator(execType types.ExecutorType, validator TaskValidator) {
	taskValidationFactory[execType] = validator
}

func GetTaskValidator(execType types.ExecutorType) (TaskValidator, error) {
	validator, exists := taskValidationFactory[execType]
	if !exists {
		return nil, errors.New(fmt.Sprintf("task validator for executor of type %s does not exist", execType))
	}
	return validator, nil
}
