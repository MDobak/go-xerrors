package xerrors

import (
	"fmt"
)

// PanicError represents an error that occurs during a panic. It is
// returned by the [Recover] and [FromRecover] functions. It provides
// access to the original panic value via the [Panic] method.
type PanicError interface {
	error

	// Panic returns the value that caused the panic.
	Panic() any
}

// Recover wraps the built-in `recover()` function, converting the
// recovered value into an error with a stack trace. The provided `fn`
// callback is only invoked when a panic occurs. The error passed to
// `fn` implements [PanicError].
//
// This function must always be used directly with the `defer`
// keyword; otherwise, it will not function correctly.
func Recover(fn func(err error)) {
	if r := recover(); r != nil {
		fn(&withStackTrace{
			err:   &panicError{panic: r},
			stack: callers(2),
		})
	}
}

// FromRecover converts the result of the built-in `recover()` into
// an error with a stack trace. The returned error implements
// [PanicError]. Returns nil if `r` is nil.
//
// This function must be called in the same function as `recover()`
// to ensure the stack trace is accurate.
func FromRecover(r any) error {
	if r == nil {
		return nil
	}
	return &withStackTrace{
		err:   &panicError{panic: r},
		stack: callers(3),
	}
}

// panicError represents an error that occurs during a panic,
// constructed from the value returned by `recover()`.
type panicError struct {
	panic any
}

// Panic implements the [PanicError] interface.
func (e *panicError) Panic() any {
	return e.panic
}

// Error implements the [error] interface.
func (e *panicError) Error() string {
	return fmt.Sprintf("panic: %v", e.panic)
}
