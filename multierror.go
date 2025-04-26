package xerrors

import (
	"errors"
	"strconv"
	"strings"
)

const multiErrorErrorPrefix = "the following errors occurred: "

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
func Append(err error, errs ...error) error {
	var me multiError
	if err != nil {
		if merr, ok := err.(multiError); ok {
			for _, e := range merr {
				if e != nil {
					me = append(me, e)
				}
			}
		} else {
			me = append(me, err)
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
	s.WriteString(multiErrorErrorPrefix)
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

// ErrorDetails implements the [ErrorDetails] interface.
func (e multiError) ErrorDetails() string {
	var s strings.Builder
	for n, err := range e.Unwrap() {
		s.WriteString(strconv.Itoa(n + 1))
		s.WriteString(". ")
		s.WriteString(indent(Sprint(err)))
	}
	return s.String()
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

// indent indents every line, except the first one, with a tab.
func indent(s string) string {
	nl := strings.HasSuffix(s, "\n")
	if nl {
		s = s[:len(s)-1]
	}
	s = strings.ReplaceAll(s, "\n", "\n\t")
	if nl {
		s += "\n"
	}
	return s
}
