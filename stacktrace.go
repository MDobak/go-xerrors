package xerrors

import (
	"fmt"
	"io"
	"runtime"
	"strconv"
	"strings"
)

const stackTraceDepth = 32

// StackTrace returns a stack trace from given error or the first stack trace
// from the wrapped errors.
func StackTrace(err error) Callers {
	for err != nil {
		if e, ok := err.(StackTracer); ok {
			return e.StackTrace()
		}
		if e, ok := err.(Wrapper); ok {
			err = e.Unwrap()
			continue
		}
		break
	}
	return nil
}

// WithStackTrace adds a stack trace to the error at the point it was called.
// The skip argument is the number of stack frames to skip.
//
// This function is useful when you want to skip the first few frames in a
// stack trace. To add a stack trace to a sentinel error, use the New function.
//
// If err is nil, then nil is returned.
func WithStackTrace(err error, skip int) error {
	if err == nil {
		return nil
	}
	return &withStackTrace{
		err:   err,
		stack: callers(skip + 1),
	}
}

// withStackTrace adds a stack trace to en error.
type withStackTrace struct {
	err   error
	stack Callers
}

// Error implements the error interface.
func (e *withStackTrace) Error() string {
	return e.err.Error()
}

// ErrorDetails implements the DetailedError interface.
func (e *withStackTrace) ErrorDetails() string {
	return e.stack.String()
}

// Unwrap implements the Wrapper interface.
func (e *withStackTrace) Unwrap() error {
	return e.err
}

// StackTrace implements the StackTracer interface.
func (e *withStackTrace) StackTrace() Callers {
	return e.stack
}

type Frame struct {
	File     string
	Line     int
	Function string
}

// String implements the fmt.Stringer interface.
func (f Frame) String() string {
	s := &strings.Builder{}
	f.writeFrame(s)
	return s.String()
}

// Format implements the fmt.Formatter interface.
//
// The verbs:
//
// 	%s	function, file and line number in a single line
// 	%f	filename
// 	%d	line number
// 	%n	function name, the plus flag adds a package name
// 	%v	same as %s, the plus or hash flags print struct details
// 	%q	a double-quoted Go string with same contents as %s
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

func (f Frame) writeFrame(w io.Writer) {
	io.WriteString(w, "\tat ")
	io.WriteString(w, shortname(f.Function))
	io.WriteString(w, " (")
	io.WriteString(w, f.File)
	io.WriteString(w, ":")
	io.WriteString(w, strconv.Itoa(f.Line))
	io.WriteString(w, ")")
}

// Callers is a list of program counters returned by the runtime.Callers.
type Callers []uintptr

// Frames returns a slice of structures with a function/file/line information.
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

// String implements the fmt.Stringer interface.
func (c Callers) String() string {
	s := &strings.Builder{}
	c.writeTrace(s)
	return s.String()
}

// Format implements the fmt.Formatter interface.
//
// The verbs:
//
// 	%s	a stack trace
// 	%v	same as %s, the plus or hash flags print struct details
// 	%q	a double-quoted Go string with same contents as %s
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

func (c Callers) writeTrace(w io.Writer) {
	frames := c.Frames()
	for _, frame := range frames {
		frame.writeFrame(w)
		io.WriteString(w, "\n")
	}
}

func callers(skip int) Callers {
	b := make([]uintptr, stackTraceDepth)
	l := runtime.Callers(skip+2, b[:])
	return b[:l]
}

func shortname(name string) string {
	i := strings.LastIndex(name, "/")
	return name[i+1:]
}
