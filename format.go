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

// Print formats an error and displays it on stderr.
//
// If the error implements the DetailedError interface, the result from the
// ErrorDetails method is used for each wrapped error, otherwise the standard
// Error method is used. A formatted error can be multi-line and always ends
// with a newline.
func Print(err error) {
	fprint(errWriter, err)
}

// Sprint formats an error and returns it as a string.
//
// If the error implements the DetailedError interface, the result from the
// ErrorDetails method is used for each wrapped error, otherwise the standard
// Error method is used. A formatted error can be multi-line and always ends
// with a newline.
func Sprint(err error) string {
	s := &strings.Builder{}
	fprint(s, err)
	return s.String()
}

// Fprint formats an error.
//
// If the error implements the DetailedError interface, the result from the
// ErrorDetails method is used for each wrapped error, otherwise the standard
// Error method is used. A formatted error can be multi-line and always ends
// with a newline.
func Fprint(w io.Writer, err error) (int, error) {
	return fprint(w, err)
}

func fprint(w io.Writer, e error) (n int, err error) {
	const firstErrorPrefix = "Error: "
	const previousErrorPrefix = "Previous error: "
	b := &bytes.Buffer{}
	f := true
	for e != nil {
		switch terr := e.(type) {
		case DetailedError:
			if f {
				b.WriteString(firstErrorPrefix)
			} else {
				b.WriteString(previousErrorPrefix)
			}
			b.WriteString(terr.Error())
			b.WriteByte('\n')
			b.WriteString(terr.ErrorDetails())
		default:
			// If an error does not implement the DetailedError interface,
			// then the Error() method will print all errors separated
			// with ":", so there is no need to render each error other than
			// the first one.
			if f {
				b.WriteString(firstErrorPrefix)
				b.WriteString(terr.Error())
				b.WriteByte('\n')
			}
		}
		f = false
		if we, ok := e.(Wrapper); ok {
			e = we.Unwrap()
			continue
		}
		break
	}
	return w.Write(b.Bytes())
}

func format(s fmt.State, verb rune, v interface{}) {
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
