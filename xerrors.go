package xerrors

import (
	"fmt"
)

// Wrapper provides context around another error.
type Wrapper interface {
	error
	Unwrap() error
}

// StackTracer provides a stack trace for an error.
type StackTracer interface {
	error
	StackTrace() Callers
}

type MultiError interface {
	error
	Errors() []error
}

// DetailedError provides extended information about an error.
// The ErrorDetails method returns a longer, multi-line description of
// the error. It must always end with a new line.
type DetailedError interface {
	error
	ErrorDetails() string
}

// messageError is the simplest possible error which contains only
// a string message.
type messageError struct {
	msg string
}

// Error implements the error interface.
func (e *messageError) Error() string {
	return e.msg
}

// Message creates a simple error with the given message. It does not record
// a stack trace. Each call returns a distinct error value even if the
// message is identical.
func Message(msg string) error {
	return &messageError{msg: msg}
}

// New creates a new error from the given value and records a stack trace at
// the point it was called. If multiple values are provided, then each error
// is wrapped by the previous error. Calling New(a, b, c), where a, b, and c
// are errors, is equivalent to calling New(WithWrapper(WithWrapper(a, b), c)).
//
// Values are converted to errors according to the following rules:
//
// - If a value is an error, it will be used as is.
//
// - If a value is a string, then new error with a given string as a message
// will be created.
//
// - If a value is nil, it will be ignored.
//
// - If a value implements the fmt.Stringer interface, then a String() method
// will be used to create an error.
//
// - For other types the result of fmt.Sprint will be used to create an error.
//
// This function may be used to:
//
// - Add a stack trace to an error: New(err)
//
// - Create a message error with a stack trace: New("access denied")
//
// - Wrap an error with a message: New("access denied", io.EOF)
//
// - Wrap one error in another: New(ErrAccessDenied, io.EOF)
//
// - Add a message to a sentinel error: New(ErrReadError, "access denied")
//
// It is possible to use errors.Is function on returned error to check whether
// an error has been used in the New function.
//
// If the function is called with no arguments or all arguments are nil, it
// returns nil.
//
// To create a simple message error without a stack trace to be used as a
// sentinel error, use the Message function instead.
func New(vals ...interface{}) error {
	var errs error
	for _, val := range vals {
		if val == nil {
			continue
		}
		err := toError(val)
		if errs == nil {
			errs = err
		} else {
			errs = &withWrapper{
				wrapper: errs,
				err:     err,
			}
		}
	}
	if errs == nil {
		return nil
	}
	return &withStackTrace{
		err:   errs,
		stack: callers(1),
	}
}

func toError(val interface{}) error {
	var err error
	switch typ := val.(type) {
	case error:
		err = typ
	case string:
		err = &messageError{msg: typ}
	case fmt.Stringer:
		err = &messageError{msg: typ.String()}
	default:
		err = &messageError{msg: fmt.Sprint(val)}
	}
	return err
}
