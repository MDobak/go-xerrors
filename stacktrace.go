package xerrors

import (
	"fmt"
	"io"
	"runtime"
	"strconv"
	"strings"
)

const stackTraceDepth = 32

// StackTrace extracts the stack trace from the provided error.
// It traverses the error chain, looking for the last error that
// has a stack trace.
func StackTrace(err error) Callers {
	var callers Callers
	for err != nil {
		if e, ok := err.(interface{ StackTrace() Callers }); ok {
			callers = e.StackTrace()
		}
		if e, ok := err.(interface{ Unwrap() error }); ok {
			err = e.Unwrap()
			continue
		}
		break
	}
	return callers
}

// WithStackTrace wraps the provided error with a stack trace,
// capturing the stack at the point of the call. The `skip` argument
// specifies how many stack frames to skip.
//
// If err is nil, WithStackTrace returns nil.
func WithStackTrace(err error, skip int) error {
	if err == nil {
		return nil
	}
	return &withStackTrace{
		err:   err,
		stack: callers(skip + 1),
	}
}

// withStackTrace wraps an error with a captured stack trace.
type withStackTrace struct {
	err   error
	stack Callers
}

// Error implements the [error] interface.
func (e *withStackTrace) Error() string {
	return e.err.Error()
}

// DetailedError implements the [DetailedError] interface.
func (e *withStackTrace) DetailedError() string {
	s := &strings.Builder{}
	s.WriteString(e.err.Error())
	s.WriteString("\n")
	s.WriteString(e.stack.String())
	return s.String()
}

// Unwrap implements the Go 1.13 `Unwrap() error` method, returning
// the wrapped error.
func (e *withStackTrace) Unwrap() error {
	return e.err
}

// StackTrace returns the stack trace captured at the point of the
// error creation.
func (e *withStackTrace) StackTrace() Callers {
	return e.stack
}

// Frame represents a single stack frame with file, line, and
// function details.
type Frame struct {
	File     string
	Line     int
	Function string
}

// String implements the [fmt.Stringer] interface.
func (f Frame) String() string {
	s := &strings.Builder{}
	f.writeFrame(s)
	return s.String()
}

// Format implements the [fmt.Formatter] interface.
//
// Supported verbs:
//   - %s function, file, and line number in a single line
//   - %f filename
//   - %d line number
//   - %n function name, with '+' flag adding the package name
//   - %v same as %s; '+' or '#' flags print struct details
//   - %q double-quoted Go string, same as %s
func (f Frame) Format(s fmt.State, verb rune) {
	type _Frame Frame
	switch verb {
	case 's':
		f.writeFrame(s)
	case 'f':
		io.WriteString(s, f.File)
	case 'd':
		io.WriteString(s, strconv.Itoa(f.Line))
	case 'n':
		switch {
		case s.Flag('+'):
			io.WriteString(s, f.Function)
		default:
			io.WriteString(s, shortname(f.Function))
		}
	case 'v':
		switch {
		case s.Flag('+') || s.Flag('#'):
			format(s, verb, _Frame(f))
		default:
			f.Format(s, 's')
		}
	case 'q':
		io.WriteString(s, strconv.Quote(f.String()))
	default:
		format(s, verb, _Frame(f))
	}
}

// writeFrame writes a formatted stack frame to the given [io.Writer].
func (f Frame) writeFrame(w io.Writer) {
	io.WriteString(w, "\tat ")
	io.WriteString(w, shortname(f.Function))
	io.WriteString(w, " (")
	io.WriteString(w, f.File)
	io.WriteString(w, ":")
	io.WriteString(w, strconv.Itoa(f.Line))
	io.WriteString(w, ")")
}

// Callers represents a list of program counters from the
// [runtime.Callers] function.
type Callers []uintptr

// Frames returns a slice of [Frame] structs with function, file, and
// line information.
func (c Callers) Frames() []Frame {
	r := make([]Frame, len(c))
	f := runtime.CallersFrames(c)
	n := 0
	for {
		frame, more := f.Next()
		r[n] = Frame{
			File:     frame.File,
			Line:     frame.Line,
			Function: frame.Function,
		}
		if !more {
			break
		}
		n++
	}
	return r
}

// String implements the [fmt.Stringer] interface.
func (c Callers) String() string {
	s := &strings.Builder{}
	c.writeTrace(s)
	return s.String()
}

// Format implements the [fmt.Formatter] interface.
//
// Supported verbs:
//   - %s complete stack trace
//   - %v same as %s; '+' or '#' flags print struct details
//   - %q double-quoted Go string, same as %s
func (c Callers) Format(s fmt.State, verb rune) {
	type _Callers Callers
	switch verb {
	case 's':
		c.writeTrace(s)
	case 'v':
		switch {
		case s.Flag('+') || s.Flag('#'):
			format(s, verb, _Callers(c))
		default:
			c.Format(s, 's')
		}
	case 'q':
		io.WriteString(s, strconv.Quote(c.String()))
	default:
		format(s, verb, _Callers(c))
	}
}

// writeTrace writes the stack trace to the provided [io.Writer].
func (c Callers) writeTrace(w io.Writer) {
	frames := c.Frames()
	for _, frame := range frames {
		frame.writeFrame(w)
		io.WriteString(w, "\n")
	}
}

// callers captures the current stack trace, skipping the specified
// number of frames.
func callers(skip int) Callers {
	b := make([]uintptr, stackTraceDepth)
	l := runtime.Callers(skip+2, b[:])
	return b[:l]
}

// shortname extracts the short name of a function, removing the
// package path.
func shortname(name string) string {
	i := strings.LastIndex(name, "/")
	return name[i+1:]
}
