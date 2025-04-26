package xerrors

import (
	"fmt"
)

// ErrorDetails represents an error that provides additional details
// beyond the error message.
//
// The ErrorDetails method returns a longer, multi-line description
// of the error. It always ends with a new line.
type ErrorDetails interface {
	error
	ErrorDetails() string
}

// Message creates a simple error with the given message, without
// recording a stack trace. Each call returns a distinct error
// instance, even if the message is identical.
//
// This function is useful for creating sentinel errors, often
// referred to as "constant errors."
func Message(msg string) error {
	return &messageError{msg: msg}
}

// New creates a new error from the provided values and records a
// stack trace at the point of the call. If multiple values are
// provided, each value is wrapped by the previous one, forming a
// chain of errors.
//
// Usage examples:
//   - Add a stack trace to an existing error: New(err)
//   - Create an error with a message and a stack trace: New("access denied")
//   - Wrap an error with a message: New("access denied", io.EOF)
//   - Add context to a sentinel error: New(ErrReadError, "access denied")
//
// Conversion rules for arguments:
//   - If the value is an error, it is used as is.
//   - If the value is a string, a new error with that message is
//     created.
//   - If the value implements [fmt.Stringer], the result of
//     String() is used to create an error.
//   - If the value is nil, it is ignored.
//   - Otherwise, the result of [fmt.Sprint] is used to create an
//     error.
//
// If called with no arguments or only nil values, New returns nil.
//
// For simple errors without a stack trace, use the [Message]
// function.
func New(vals ...any) error {
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

// unwrapper represents an error that wraps another error, providing
// additional context.
type unwrapper interface {
	error
	Unwrap() error
}

// messageError represents a simple error that contains only a string
// message.
type messageError struct {
	msg string
}

// Error implements the [error] interface.
func (e *messageError) Error() string {
	return e.msg
}

func toError(val any) error {
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
