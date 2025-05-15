package xerrors

import (
	"errors"
	"strconv"
	"strings"
)

// Append appends the provided errors to an existing error or list of
// errors. If `err` is not a [multiError], it will be converted into
// one. Nil errors are ignored. It does not record a stack trace.
//
// If the resulting error list is empty, nil is returned. If the
// resulting error list contains only one error, that error is
// returned instead of the list.
//
// The returned error is compatible with Go errors, supporting
// [errors.Is], [errors.As], and the Go 1.20 `Unwrap() []error`
// method.
//
// To create a chained error, use [New], [Newf], [Join], or
// [Joinf] instead.
func Append(err error, errs ...error) error {
	var me multiError
	if err != nil {
		if mErr, ok := err.(multiError); ok {
			me = mErr
		} else {
			me = multiError{err}
		}
	}
	for _, e := range errs {
		if e != nil {
			me = append(me, e)
		}
	}
	switch len(me) {
	case 0:
		return nil
	case 1:
		return me[0]
	default:
		return me
	}
}

// multiError is a slice of errors that can be treated as a single
// error.
type multiError []error

// Error implements the [error] interface.
func (e multiError) Error() string {
	var s strings.Builder
	s.WriteString("[")
	for n, err := range e {
		s.WriteString(err.Error())
		if n < len(e)-1 {
			s.WriteString(", ")
		}
	}
	s.WriteString("]")
	return s.String()
}

// ErrorDetails returns additional details about the error for
// the [ErrorDetails] function.
func (e multiError) ErrorDetails() string {
	if len(e) == 0 {
		return ""
	}
	buf := &strings.Builder{}
	for n, err := range e.Unwrap() {
		buf.WriteString(strconv.Itoa(n + 1))
		buf.WriteString(". ")
		writeErr(buf, err)
	}
	return buf.String()
}

// Unwrap implements the Go 1.20 `Unwrap() []error` method, returning
// a slice containing all errors in the list.
func (e multiError) Unwrap() []error {
	s := make([]error, len(e))
	copy(s, e)
	return s
}

// As implements the Go 1.13 `errors.As` method, allowing type
// assertions on all errors in the list.
func (e multiError) As(target any) bool {
	for _, err := range e {
		if errors.As(err, target) {
			return true
		}
	}
	return false
}

// Is implements the Go 1.13 `errors.Is` method, allowing
// comparisons with all errors in the list.
func (e multiError) Is(target error) bool {
	for _, err := range e {
		if errors.Is(err, target) {
			return true
		}
	}
	return false
}
