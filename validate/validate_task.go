package validate

import (
	"errors"
	"fmt"
	"github.com/GateOfBabylon/enuma-elish-interpreter/types"
)

type HttpTaskValidator struct{}

func (htv *HttpTaskValidator) Validate(task *types.Task) error {
	if err := generalTaskValidation(task); err != nil {
		return err
	}
	if task.PyTaskFields != nil {
		return errors.New("HTTP task could not use other task specific fields")
	}
	if task.HttpTaskFields != nil {
		if !IsSupportedHttpMethod(task.HttpTaskFields.Method) {
			return errors.New(fmt.Sprintf("HTTP method %q is not supported", task.HttpTaskFields.Method))
		}
		if task.HttpTaskFields.Url == "" {
			return errors.New("URL is required")
		}
	}
	return nil
}

type PythonTaskValidator struct{}

func (ptv *PythonTaskValidator) Validate(task *types.Task) error {
	if err := generalTaskValidation(task); err != nil {
		return err
	}

	if task.HttpTaskFields != nil {
		return errors.New("python task could not use other task specific fields")
	}

	if task.PyTaskFields != nil {
		// One of script or script path must be different from empty string
		if task.PyTaskFields.Script == "" && task.PyTaskFields.ScriptPath == "" {
			return errors.New("script or script path is required")
		}
		if task.PyTaskFields.Script != "" && task.PyTaskFields.ScriptPath != "" {
			return errors.New("script and script path are mutually exclusive")
		}
	}
	return nil
}

func generalTaskValidation(task *types.Task) error {
	if err := validateTimeStatements(task.TimeStatements); err != nil {
		return fmt.Errorf("time statements validation failed: %w", err)
	}
	if err := validateConditionStatements(task.ConditionStatements); err != nil {
		return fmt.Errorf("condition statements validation failed: %w", err)
	}

	return nil
}

func validateTimeStatements(timeStatements *types.TimeStatements) error {
	if timeStatements == nil {
		return nil // TimeStatements are optional
	}

	if timeStatements.Delay < 0 {
		return errors.New("delay must be a non-negative integer")
	}
	if timeStatements.Timeout < 0 {
		return errors.New("timeout must be a non-negative integer")
	}
	if timeStatements.Retry < 0 {
		return errors.New("retry must be a non-negative integer")
	}

	return nil
}

func validateConditionStatements(conditionStatements *types.ConditionStatements) error {
	if conditionStatements == nil {
		return nil // ConditionStatements are optional
	}

	if conditionStatements.Condition != "" && !isValidCondition(conditionStatements.Condition) {
		return errors.New(fmt.Sprintf("invalid condition syntax: %s", conditionStatements.Condition))
	}

	if conditionStatements.Pick != nil {
		if err := validatePickStatement(conditionStatements.Pick); err != nil {
			return errors.New(fmt.Sprintf("pick statement validation failed: %w", err))
		}
	}

	if conditionStatements.Parallel != nil {
		if len(*conditionStatements.Parallel) == 0 {
			return errors.New("parallel tasks cannot be empty")
		}
		for _, parallelTask := range *conditionStatements.Parallel {
			if err := generalTaskValidation(&parallelTask); err != nil {
				return fmt.Errorf("parallel task validation failed: %w", err)
			}
		}
	}

	return nil
}

func validatePickStatement(pick *types.PickStatement) error {
	if len(pick.IfStatement) == 0 {
		return errors.New("at least one if statement is required in a pick block")
	}

	for _, ifStmt := range pick.IfStatement {
		if ifStmt.Try == "" {
			return errors.New("try condition is mandatory for if statements")
		}
		if ifStmt.Task == nil {
			return errors.New("task is mandatory for if statements")
		}
		if err := generalTaskValidation(ifStmt.Task); err != nil {
			return fmt.Errorf("if statement task validation failed: %w", err)
		}
	}

	if pick.Else != nil {
		if err := generalTaskValidation(pick.Else); err != nil {
			return fmt.Errorf("else block validation failed: %w", err)
		}
	}

	return nil
}

func isValidCondition(condition string) bool {
	// Validate condition syntax (e.g., logical expressions like ${{status == 'ready'}})
	// Example placeholder implementation
	return len(condition) > 0
}

func init() {
	RegisterTaskValidator(types.HTTP, &HttpTaskValidator{})
	RegisterTaskValidator(types.PYTHON, &PythonTaskValidator{})
}

func ValidateTask(task *types.Task, executorType types.ExecutorType) error {
	validator, err := GetTaskValidator(executorType)
	if err != nil {
		return err
	}
	return validator.Validate(task)
}
