package xerrors

import (
	"fmt"
)

// Recover wraps the recover() built-in and converts a value returned by it to
// an error with a stack trace. The fn callback will be invoked only during
// panicking.
//
// This function must always be used *directly* with the "defer" keyword.
// Otherwise, it will not work.
func Recover(fn func(err error)) {
	if r := recover(); r != nil {
		fn(&withStackTrace{
			err:   &panicError{panic: r},
			stack: callers(2),
		})
	}
}

// FromRecover takes the result of the recover() built-in and converts it to
// an error with a stack trace.
//
// This function must be invoked in the same function as recover(), otherwise
// the returned stack trace will not be correct.
func FromRecover(r interface{}) error {
	if r == nil {
		return nil
	}
	return &withStackTrace{
		err:   &panicError{panic: r},
		stack: callers(3),
	}
}

// panicError is an error constructed from a value returned by the recover()
// built-in during panicking.
type panicError struct {
	panic interface{}
}

// Panic returns the value from the recover() function.
func (e *panicError) Panic() interface{} {
	return e.panic
}

// Error implements the error interface.
func (e *panicError) Error() string {
	return fmt.Sprintf("panic: %v", e.panic)
}
