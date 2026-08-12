package xerrors

import (
	"fmt"
	"io"
	"runtime"
	"strconv"
	"strings"
)

// DefaultCallersFormatter is the default formatter for [Callers].
var DefaultCallersFormatter = func(c Callers, w io.Writer) {
	for _, frame := range c.Frames() {
		io.WriteString(w, "at ")
		frame.writeFrame(w)
		io.WriteString(w, "\n")
	}
}

// DefaultFrameFormatter is the default formatter for [Frame].
var DefaultFrameFormatter = func(f Frame, w io.Writer) {
	io.WriteString(w, shortname(f.Function))
	io.WriteString(w, " (")
	io.WriteString(w, f.File)
	io.WriteString(w, ":")
	io.WriteString(w, strconv.Itoa(f.Line))
	io.WriteString(w, ")")
}

const stackTraceDepth = 128

// StackTrace extracts the stack trace from the provided error.
// It traverses the error chain and returns the stack trace of the
// innermost error that has one. It returns nil if no error in the
// chain has a stack trace.
func StackTrace(err error) Callers {
	var callers Callers
	for err != nil {
		if st, ok := err.(interface{ StackTrace() Callers }); ok {
			callers = st.StackTrace()
		}
		if wErr, ok := err.(interface{ Unwrap() error }); ok {
			err = wErr.Unwrap()
			continue
		}
		break
	}
	return callers
}

// WithStackTrace wraps the provided error with a stack trace,
// capturing the stack at the point of the call. The `skip` argument
// specifies how many stack frames to skip: 0 starts the stack trace
// at the caller of WithStackTrace, 1 at its caller, and so on.
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

// ErrorDetails implements the [DetailedError] interface.
func (e *withStackTrace) ErrorDetails() string {
	return e.stack.String()
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
//   - %s function, file, and line number on a single line
//   - %f file path
//   - %d line number
//   - %n function name; the '+' flag prints the package path as well
//   - %v same as %s; the '+' and '#' flags print the struct fields
//   - %q the result of %s as a double-quoted Go string
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
	DefaultFrameFormatter(f, w)
}

// Callers is a stack trace represented as a list of program counters,
// as returned by [runtime.Callers].
type Callers []uintptr

// Frames returns a slice of [Frame] structs with function, file, and
// line information. It returns nil if the stack trace is empty.
func (c Callers) Frames() []Frame {
	if len(c) == 0 {
		return nil
	}
	r := make([]Frame, 0, len(c))
	f := runtime.CallersFrames(c)
	for {
		frame, more := f.Next()
		r = append(r, Frame{
			File:     frame.File,
			Line:     frame.Line,
			Function: frame.Function,
		})
		if !more {
			break
		}
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
//   - %s the complete stack trace, one frame per line
//   - %v same as %s; the '+' and '#' flags print the raw program counters
//   - %q the result of %s as a double-quoted Go string
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
	DefaultCallersFormatter(c, w)
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
