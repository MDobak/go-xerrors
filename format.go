package xerrors

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

var errWriter io.Writer = os.Stderr

// Print writes a formatted error to stderr.
//
// If the error implements the [DetailedError] interface and returns
// a non-empty string, the returned details are added to each error
// in the chain.
//
// The formatted error can span multiple lines and always ends with
// a newline.
func Print(err error) {
	buf := &strings.Builder{}
	writeErr(buf, err)
	errWriter.Write([]byte(buf.String()))
}

// Sprint returns a formatted error as a string.
//
// If the error implements the [DetailedError] interface and returns
// a non-empty string, the returned details are added to each error
// in the chain.
//
// The formatted error can span multiple lines and always ends with
// a newline.
func Sprint(err error) string {
	buf := &strings.Builder{}
	writeErr(buf, err)
	return buf.String()
}

// Fprint writes a formatted error to the provided [io.Writer].
//
// If the error implements the [DetailedError] interface and returns
// a non-empty string, the returned details are added to each error
// in the chain.
//
// The formatted error can span multiple lines and always ends with
// a newline.
func Fprint(w io.Writer, err error) (int, error) {
	buf := &strings.Builder{}
	writeErr(buf, err)
	return w.Write([]byte(buf.String()))
}

// writeErr writes a formatted error to the provided strings.Builder.
func writeErr(buf *strings.Builder, err error) {
	const firstErrorPrefix = "Error: "
	const previousErrorPrefix = "Previous error: "
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
			buf.WriteString("\n\t")
			buf.WriteString(indent(errDetails))
			if !strings.HasSuffix(errDetails, "\n") {
				buf.WriteByte('\n')
			}
		} else {
			// If an error does not contain any details, do not print
			// it, except for the first one. This is to avoid printing
			// every wrapped error on a single line.
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
}

// format is a helper function that formats a value according to the provided
// format state and verb.
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
