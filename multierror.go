package xerrors

import (
	"errors"
	"strconv"
	"strings"
)

// Append adds more errors to an existing list of errors. If err is not a list
// of errors, then it will be converted into a list. It does not record
// a stack trace.
//
// The returned list of errors is compatible with Go 1.13 errors, and it
// supports the errors.Is and errors.As methods. However, the errors.Unwrap
// method is not supported.
//
// If err is nil and no additional errors are given, nil is returned.
//
// Append is not thread-safe.
func Append(err error, errs ...error) error {
	if err == nil && len(errs) == 0 {
		return nil
	}
	switch errTyp := err.(type) {
	case multiError:
		return append(errTyp, errs...)
	default:
		var me multiError
		if err != nil {
			me = multiError{err}
		}
		for _, e := range errs {
			if e != nil {
				me = append(me, e)
			}
		}
		if len(me) == 0 {
			return nil
		}
		return me
	}
}

const multiErrorErrorPrefix = "the following errors occurred: "

// multiError is a slice of errors that can be used as a single error.
type multiError []error

// Error implements the error interface.
func (e multiError) Error() string {
	s := &strings.Builder{}
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

// ErrorDetails implements the DetailedError interface.
func (e multiError) ErrorDetails() string {
	s := &strings.Builder{}
	for n, err := range e.Errors() {
		s.WriteString(strconv.Itoa(n + 1))
		s.WriteString(". ")
		s.WriteString(indent(Sprint(err)))
	}
	return s.String()
}

// Errors implements the MultiError interface.
func (e multiError) Errors() []error {
	s := make([]error, len(e))
	for i, err := range e {
		s[i] = err
	}
	return s
}

func (e multiError) As(target interface{}) bool {
	for _, err := range e {
		if errors.As(err, target) {
			return true
		}
	}
	return false
}

func (e multiError) Is(target error) bool {
	for _, err := range e {
		if errors.Is(err, target) {
			return true
		}
	}
	return false
}

// indent idents every line, except the first one, with tab.
func indent(s string) string {
	end := ""
	if strings.HasSuffix(s, "\n") {
		end = "\n"
		s = s[:len(s)-1]
	}
	return strings.ReplaceAll(s, "\n", "\n\t") + end
}
