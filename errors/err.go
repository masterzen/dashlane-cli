package errors

import (
	"fmt"
	"os"
)

// ExitError is an `error` with an `exitCode`
type ExitError interface {
	error
	ExitCode() int
}

// ExitCodeError allows for a command to return an error with a given exit code
// in which case our main will return this exitCode to the OS
type ExitCodeError struct {
	exitCode int
	message  string
}

// NewExitCodeError makes a new *ExitCodeError
func NewExitCodeError(message string, exitCode int) *ExitCodeError {
	return &ExitCodeError{
		exitCode: exitCode,
		message:  message,
	}
}

// NewExitCodeFromError makes a new *ExitCodeError
func NewExitCodeFromError(err error, exitCode int) *ExitCodeError {
	return &ExitCodeError{
		exitCode: exitCode,
		message:  err.Error(),
	}
}

// Error returns this error string message
func (e *ExitCodeError) Error() string {
	return e.message
}

// ExitCode returns this error exit code,
func (e *ExitCodeError) ExitCode() int {
	return e.exitCode
}

// ExitOnExitCodeError really exits in case of error and return an exit code
func ExitOnExitCodeError(err error) {
	if err == nil {
		return
	}

	if codeError, ok := err.(ExitError); ok {
		if err.Error() != "" {
			fmt.Println(err)
		}
		os.Exit(codeError.ExitCode())
		return
	}
}
