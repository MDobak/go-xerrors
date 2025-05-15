package xerrors

import (
	"errors"
	"strings"
)

// withWrapper wraps an error with another error.
//
// It is intended to be build error chains, e.g. if we have a
// following error chain: `err1: err2: err3`, the wrapper is `err1`,
// and the err is another withWrapper containing `err2` and `err3`.
type withWrapper struct {
	wrapper error  // wrapper is the error that wraps the next error in the chain, may be nil
	err     error  // err is the next error in the chain, must not be nil
	msg     string // msg overwrites the error message, if set
}

// Error implements the [error] interface.
func (e *withWrapper) Error() string {
	if e.msg != "" {
		return e.msg
	}
	s := &strings.Builder{}
	if e.wrapper != nil {
		s.WriteString(e.wrapper.Error())
		s.WriteString(": ")
	}
	s.WriteString(e.err.Error())
	return s.String()
}

// ErrorDetails implements the [DetailedError] interface.
func (e *withWrapper) ErrorDetails() string {
	err := e.wrapper
	for err != nil {
		if dErr, ok := err.(DetailedError); ok {
			return dErr.ErrorDetails()
		}
		if wErr, ok := err.(interface{ Unwrap() error }); ok {
			err = wErr.Unwrap()
			continue
		}
		break
	}
	return ""
}

// Unwrap implements the Go 1.13 `Unwrap() error` method, returning
// the wrapped error.
//
// Since withWrapper represents a chain of errors, the Unwrap method
// returns the next error in the chain, not both the wrapper and the error.
func (e *withWrapper) Unwrap() error {
	return e.err
}

// As implements the Go 1.13 `errors.As` method, allowing type
// assertions on all errors in the list.
func (e *withWrapper) As(target any) bool {
	return errors.As(e.wrapper, target) || errors.As(e.err, target)
}

// Is implements the Go 1.13 `errors.Is` method, allowing
// comparisons with all errors in the list.
func (e *withWrapper) Is(target error) bool {
	return errors.Is(e.wrapper, target) || errors.Is(e.err, target)
}
