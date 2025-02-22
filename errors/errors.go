package errors

import (
	"errors"
	"fmt"
)

var (
	// copy these functions here to avoid package name conflict
	New = errors.New
)

type UserFixableError struct {
	Err          error
	Instructions string
}

func NewUserFixableError(err error, instructions string) UserFixableError {
	return UserFixableError{
		Err:          err,
		Instructions: instructions,
	}
}

func (e UserFixableError) Error() string {
	return e.Err.Error()
}

func PrintUserFixableErrorInstruction(err error) {
	var configErr UserFixableError
	if errors.As(err, &configErr) {
		fmt.Println(configErr.Instructions)
	}
}
