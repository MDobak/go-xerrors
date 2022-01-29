package xerrors

import (
	"errors"
	"strings"
)

// WithWrapper wraps err with wrapper.
//
// The error used as wrapper should be a simple error, preferably a sentinel
// error. This is because details such as the wrapper's stack trace are ignored.
//
// The Unwrap method will unwrap only err but errors.Is, errors.As works with
// both of the errors.
//
// If wrapper is nil, then err is returned.
// If err is nil, then nil is returned.
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

// Error implements the error interface.
func (e *withWrapper) Error() string {
	s := &strings.Builder{}
	s.WriteString(e.wrapper.Error())
	s.WriteString(": ")
	s.WriteString(e.err.Error())
	return s.String()
}

// Unwrap implements the Wrapper interface.
func (e *withWrapper) Unwrap() error {
	return e.err
}

func (e *withWrapper) As(target interface{}) bool {
	return errors.As(e.wrapper, target) || errors.As(e.err, target)
}

func (e *withWrapper) Is(target error) bool {
	return errors.Is(e.wrapper, target) || errors.Is(e.err, target)
}
