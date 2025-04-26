package xerrors

import (
	"errors"
	"fmt"
	"regexp"
	"testing"
)

func TestRecover(t *testing.T) {
	tests := []struct {
		panic any
		want  string
	}{
		{panic: nil, want: ""},
		{panic: "foo", want: "panic: foo"},
		{panic: 42, want: "panic: 42"},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			handled := false
			defer func() {
				if tt.panic != nil && !handled {
					t.Errorf("Recover(): callback was not called during panicking")
				}
				if tt.panic == nil && handled {
					t.Errorf("Recover(): callback was called without panickng")
				}
			}()
			defer Recover(func(got error) {
				handled = true
				if got.Error() != tt.want {
					t.Errorf("Recover(): got: %q, want %q", got, tt.want)
				}
				st := StackTrace(got)
				if len(st) == 0 {
					t.Errorf("Recover(): created error must contain a stack trace")
				}
				if len(st) > 0 && shortname(st.Frames()[0].Function) != "go-xerrors.TestRecover.func1" {
					t.Errorf("Recover(): the first frame of stack trace must start at xerrors.TestRecover.func1")
				}
				panicErr := &panicError{}
				if errors.As(got, &panicErr); panicErr.Panic() != tt.panic {
					t.Errorf("Recover(): the value returned by Panic method must be the same as the value used to invoke panic")
				}
			})
			if tt.panic != nil {
				panic(tt.panic)
			}
		})
	}
}

func TestFromRecover(t *testing.T) {
	tests := []struct {
		panic any
		want  string
	}{
		{panic: nil, want: ""},
		{panic: "foo", want: "panic: foo"},
		{panic: 42, want: "panic: 42"},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			defer func() {
				got := FromRecover(recover())
				if tt.panic == nil {
					if got != nil {
						t.Errorf("FromRecover(nil): got: %q, want %q", got, tt.want)
					}
				} else {
					if got.Error() != tt.want {
						t.Errorf("FromRecover(): got: %q, want %q", got, tt.want)
					}
					st := StackTrace(got)
					if len(st) == 0 {
						t.Errorf("FromRecover(): created error must contain a stack trace")
					}
					if len(st) > 0 && shortname(st.Frames()[0].Function) != "go-xerrors.TestFromRecover.func1" {
						t.Errorf("FromRecover(): the first frame of stack trace must start at xerrors.TestFromRecover.func1")
					}
					panicErr := &panicError{}
					if errors.As(got, &panicErr); panicErr.Panic() != tt.panic {
						t.Errorf("FromRecover(): the value returned by Panic method must be the same as the value used to invoke panic")
					}
				}
			}()
			panic(tt.panic)
		})
	}
}

func TestPanicErrorFormat(t *testing.T) {
	tests := []struct {
		format string
		want   string
		regexp bool
	}{
		{format: "%s", want: `panic: foo`},
		{format: "%v", want: `panic: foo`},
		{format: "%q", want: `"panic: foo"`},
		{format: "%+q", want: `"panic: foo"`},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			defer Recover(func(err error) {
				if tt.regexp {
					got := fmt.Sprintf(tt.format, err)
					if match, _ := regexp.MatchString(tt.want, got); !match {
						t.Errorf("fmt.Sprtinf(%q, &panicError{...}): %q does not match %q", tt.format, got, tt.want)
					}
				} else {
					if got := fmt.Sprintf(tt.format, err); got != tt.want {
						t.Errorf("fmt.Sprtinf(%q, &panicError{...}): got: %q, want: %q", tt.format, got, tt.want)
					}
				}
			})
			panic("foo")
		})
	}
}
