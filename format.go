package xerrors

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

var errWriter io.Writer = os.Stderr

// Print writes a formatted error to stderr.
//
// If the error implements the [DetailedError] interface, the result
// of [DetailedError] is used for each wrapped error. Otherwise, the
// standard Error method is used. The formatted error can span
// multiple lines and always ends with a newline.
func Print(err error) {
	fprint(errWriter, err)
}

// Sprint returns a formatted error as a string.
//
// If the error implements the [DetailedError] interface, the result
// of [DetailedError] is used for each wrapped error. Otherwise, the
// standard Error method is used. The formatted error can span
// multiple lines and always ends with a newline.
func Sprint(err error) string {
	s := &strings.Builder{}
	fprint(s, err)
	return s.String()
}

// Fprint writes a formatted error to the provided [io.Writer].
//
// If the error implements the [DetailedError] interface, the result
// of [DetailedError] is used for each wrapped error. Otherwise, the
// standard Error method is used. The formatted error can span
// multiple lines and always ends with a newline.
func Fprint(w io.Writer, err error) (int, error) {
	return fprint(w, err)
}

// fprint is a helper function that writes a formatted error to the
// given [io.Writer].
//
// This function prints all errors in the chain that implement the
// [DetailedError] interface. The first error is printed using the
// standard Error method if it does not implement [DetailedError].
func fprint(w io.Writer, err error) (int, error) {
	const firstErrorPrefix = "Error: "
	const previousErrorPrefix = "Previous error: "
	var buf bytes.Buffer
	first := true
	for err != nil {
		errMsg := err.Error()
		errDetails := ""
		if dErr, ok := err.(DetailedError); ok {
			errDetails = dErr.ErrorDetails()
		}
		if errDetails != "" {
			if first {
				buf.WriteString(firstErrorPrefix)
			} else {
				buf.WriteString(previousErrorPrefix)
			}
			buf.WriteString(errMsg)
			buf.WriteString("\n")
			buf.WriteString(errDetails)
		} else {
			// If an error does not have any details, then the Error() method
			// should print all errors separated with ":", so there is no need
			// to render each error other than the first one.
			if first {
				buf.WriteString(firstErrorPrefix)
				buf.WriteString(errMsg)
				buf.WriteByte('\n')
			}
		}
		first = false
		if wErr, ok := err.(interface{ Unwrap() error }); ok {
			err = wErr.Unwrap()
			continue
		}
		break
	}
	return w.Write(buf.Bytes())
}

// format is a helper function to format custom types when
// implementing [fmt.Formatter].
func format(s fmt.State, verb rune, v any) {
	f := []rune{'%'}
	for _, c := range []int{'-', '+', '#', ' ', '0'} {
		if s.Flag(c) {
			f = append(f, rune(c))
		}
	}
	if w, ok := s.Width(); ok {
		f = append(f, []rune(strconv.Itoa(w))...)
	}
	if p, ok := s.Precision(); ok {
		f = append(f, '.')
		f = append(f, []rune(strconv.Itoa(p))...)
	}
	f = append(f, verb)
	fmt.Fprintf(s, string(f), v)
}
