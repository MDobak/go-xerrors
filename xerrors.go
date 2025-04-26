package xerrors

import (
	"fmt"
)

// DetailedError represents an error that provides additional details
// beyond the error message.
//
// The DetailedError method returns a longer, multi-line description
// of the error. It always ends with a new line.
type DetailedError interface {
	error
	DetailedError() string
}

// Message creates a simple error with the given message, without
// recording a stack trace. Each call returns a distinct error
// instance, even if the message is identical.
//
// This function is useful for creating sentinel errors, often
// referred to as "constant errors."
//
// To create an error with a stack trace, use [New] or [Newf]
// instead.
func Message(msg string) error {
	return &messageError{msg: msg}
}

// Messagef creates a simple error with a formatted message,
// without recording a stack trace. The format string follows the
// conventions of [fmt.Sprintf]. Each call returns a distinct error
// instance, even if the message is identical.
//
// This function is useful for creating sentinel errors, often
// referred to as "constant errors."
//
// To create an error with a stack trace, use [New] or [Newf]
// instead.
func Messagef(format string, args ...any) error {
	return &messageError{msg: fmt.Sprintf(format, args...)}
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
// To create a sentinel error, use [Message] or [Messagef] instead.
func New(vals ...any) error {
	err := Join(vals...)
	if err == nil {
		return nil
	}
	return &withStackTrace{
		err:   err,
		stack: callers(1),
	}
}

// Newf creates a new error with a formatted message and records a
// stack trace at the point of the call. The format string follows
// the conventions of [fmt.Errorf].
//
// Unlike errors created by [fmt.Errorf], the Unwrap method on the
// returned error yields the next wrapped error, not a slice of errors,
// since this function is intended for creating linear error chains.
//
// To create a sentinel error, use [Message] or [Messagef] instead.
func Newf(format string, args ...any) error {
	return &withStackTrace{
		err:   Joinf(format, args...),
		stack: callers(1),
	}
}

// Join joins multiple values into a single error, forming a chain
// of errors.
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
// If called with no arguments or only nil values, Join returns nil.
//
// To create a multi-error instead of an error chain, use [Append].
func Join(vals ...any) error {
	var wErr error
	for i := len(vals) - 1; i >= 0; i-- {
		if vals[i] == nil {
			continue
		}
		err := toError(vals[i])
		if wErr == nil {
			wErr = err
			continue
		}
		wErr = &withWrapper{
			wrapper: err,
			err:     wErr,
		}
	}
	return wErr
}

// Joinf joins multiple values into a single error with a formatted
// message, forming an error chain. The format string follows the
// conventions of [fmt.Errorf].
//
// Unlike errors created by [fmt.Errorf], the Unwrap method on the
// returned error yields the next wrapped error, not a slice of errors,
// since this function is intended for creating linear error chains.
//
// To create a multi-error instead of an error chain, use [Append].
func Joinf(format string, args ...any) error {
	err := fmt.Errorf(format, args...)
	switch u := err.(type) {
	case interface {
		Unwrap() error
	}:
		return &withWrapper{
			err: u.Unwrap(),
			msg: err.Error(),
		}
	case interface {
		Unwrap() []error
	}:
		var wErr error
		errs := u.Unwrap()
		for i := len(errs) - 1; i >= 0; i-- {
			if errs[i] == nil {
				continue
			}
			if wErr == nil {
				wErr = errs[i]
				continue
			}
			wErr = &withWrapper{
				wrapper: errs[i],
				err:     wErr,
			}
		}
		// Because the formatted message may not follow the "err1: err2: err3"
		// pattern, we set the msg field to overwrite the wrapper's message.
		if wErr, ok := wErr.(*withWrapper); ok {
			wErr.msg = err.Error()
			return wErr
		}
		return &withWrapper{
			err: wErr,
			msg: err.Error(),
		}
	default:
		return &messageError{msg: err.Error()}
	}
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
