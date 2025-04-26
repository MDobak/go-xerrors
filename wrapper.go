package xerrors

import (
	"errors"
	"strings"
)

// WithWrapper wraps `err` with a `wrapper` error.
//
// The `wrapper` should generally be a simple, sentinel error, as
// details like its stack trace are ignored. The `Unwrap` method
// will only unwrap `err`, but [errors.Is] and [errors.As] work
// with both `wrapper` and `err`.
//
// If `wrapper` is nil, `err` is returned. If `err` is nil,
// WithWrapper returns nil.
func WithWrapper(wrapper error, err error) error {
	if err == nil {
		return nil
	}
	if wrapper == nil {
		return err
	}
	return &withWrapper{
		wrapper: wrapper,
		err:     err,
	}
}

// withWrapper wraps an error with another error.
type withWrapper struct {
	wrapper error
	err     error
}

// Error implements the [error] interface.
func (e *withWrapper) Error() string {
	s := &strings.Builder{}
	s.WriteString(e.wrapper.Error())
	s.WriteString(": ")
	s.WriteString(e.err.Error())
	return s.String()
}

// Unwrap implements the Go 1.13 `Unwrap() []error` method, returning
// the wrapped error.
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
