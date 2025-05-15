package xerrors

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"regexp"
	"testing"
)

func TestWithStackTrace(t *testing.T) {
	tests := []struct {
		err  error
		want string
	}{
		{err: Message("foo"), want: "foo"},
		{err: io.EOF, want: "EOF"},
		{err: nil, want: ""},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			err := WithStackTrace(tt.err, 0)
			if tt.err == nil {
				if err != nil {
					t.Errorf("WithStackTrace(nil): must return nil")
				}
			} else {
				if got := err.Error(); got != tt.want {
					t.Errorf("WithStackTrace(%#v).Error(): got: %q, want: %q", tt.err, got, tt.want)
				}
				if len(StackTrace(err)) == 0 {
					t.Errorf("WithStackTrace(%#v): returned error must contain a stack trace", tt.err)
				}
				if tt.err != nil {
					if !errors.Is(err, tt.err) {
						t.Errorf("errors.Is(WithStackTrace(%#v), err): must return true", tt.err)
					}
					if !errors.As(err, reflect.New(reflect.TypeOf(tt.err)).Interface()) {
						t.Errorf("errors.As(WithStackTrace(%#v), err): must return true", tt.err)
					}
				}
			}
		})
	}
}

func TestWithStackTraceFormat(t *testing.T) {
	tests := []struct {
		format string
		err    error
		skip   int
		want   string
		regexp bool
	}{
		{format: "%s", err: Message(""), want: ``},
		{format: "%s", err: New("foo"), want: `foo`},
		{format: "%s", err: io.EOF, want: `EOF`},
		{format: "%v", err: New("foo"), want: `foo`},
		{format: "%q", err: Message("foo"), want: `"foo"`},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			var err error
			func() {
				// We are running this in a closure to test stack trace frames
				// skipping.
				err = WithStackTrace(tt.err, tt.skip)
			}()
			if got := fmt.Sprintf(tt.format, err); got != tt.want {
				t.Errorf("fmt.Sprtinf(%q, WithStackTrace(%q)): got: %q, want: %q", tt.format, tt.err, got, tt.want)
			}
		})
	}
}

func TestFrameFormat(t *testing.T) {
	frame := Frame{
		File:     "file",
		Line:     42,
		Function: "package/function",
	}
	tests := []struct {
		format string
		want   string
		regexp bool
	}{
		{format: "%s", want: "function (file:42)"},
		{format: "%f", want: "file"},
		{format: "%d", want: "42"},
		{format: "%n", want: "function"},
		{format: "%+n", want: "package/function"},
		{format: "%+n", want: "package/function"},
		{format: "%v", want: "function (file:42)"},
		{format: "%+v", want: "{File:file Line:42 Function:package/function}"},
		{format: "%#v", want: "xerrors._Frame{File:\"file\", Line:42, Function:\"package/function\"}"},
		{format: "%q", want: "\"function (file:42)\""},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			if got := fmt.Sprintf(tt.format, frame); got != tt.want {
				t.Errorf("fmt.Sprtinf(%q, %#v): got: %q, want: %q", tt.format, frame, got, tt.want)
			}
		})
	}
}

func TestCallersFormat(t *testing.T) {
	callers := callers(0)
	tests := []struct {
		format string
		want   string
	}{
		{format: "%s", want: `^at .*(\nat .*)+\n$`},
		{format: "%v", want: `^at .*(\nat .*)+\n$`},
		{format: "%+v", want: `\[([0-9 ])+\]`},
		{format: "%#v", want: `^xerrors\._Callers\{(0x[a-f0-9]+, )*(0x[a-f0-9]+)\}$`},
		{format: "%q", want: `^"at .*(\\nat .*)+\\n"$`},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			got := fmt.Sprintf(tt.format, callers)
			if match, _ := regexp.MatchString(tt.want, got); !match {
				t.Errorf("fmt.Sprtinf(%q, callers(0)): %q does not match %q", tt.format, got, tt.want)
			}
		})
	}
}
