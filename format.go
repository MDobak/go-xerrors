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

// fprint is a helper function that writes the formatted error to the
// given [io.Writer].
//
// This function will print all errors in the chain, that implement the
// [DetailedError] interface, with the exception of the first error, which
// will be printed using the standard Error method if it doesn't implement
// the [DetailedError] interface.
func fprint(w io.Writer, err error) (int, error) {
	const firstErrorPrefix = "Error: "
	const previousErrorPrefix = "Previous error: "
	var buffer bytes.Buffer
	first := true
	for err != nil {
		switch tErr := err.(type) {
		case DetailedError:
			if first {
				buffer.WriteString(firstErrorPrefix)
			} else {
				buffer.WriteString(previousErrorPrefix)
			}
			buffer.WriteString(tErr.DetailedError())
		default:
			// If an error does not implement the DetailedError interface,
			// then the Error() method should print all errors separated
			// with ":", so there is no need to render each error other than
			// the first one.
			if first {
				buffer.WriteString(firstErrorPrefix)
				buffer.WriteString(tErr.Error())
				buffer.WriteByte('\n')
			}
		}
		first = false
		if wErr, ok := err.(interface{ Unwrap() error }); ok {
			err = wErr.Unwrap()
			continue
		}
		break
	}
	return w.Write(buffer.Bytes())
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
