package xerrors

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

type stringer struct{ s string }

func (s stringer) String() string {
	return s.s
}

func TestMessage(t *testing.T) {
	tests := []struct {
		val  string
		want string
	}{
		{val: "", want: ""},
		{val: "foo", want: "foo"},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			got1 := Message(tt.val)
			got2 := Message(tt.val)
			if msg := got1.Error(); msg != tt.want {
				t.Errorf("Message(%#v): got: %q, want %q", tt.val, msg, tt.want)
			}
			if len(StackTrace(got1)) != 0 {
				t.Errorf("Message(%#v): returned error must not contain a stack trace", tt.val)
			}
			if got1 == got2 {
				t.Errorf("Message(%#v): returned error must not be the same instance", tt.val)
			}
		})
	}
}

func TestMessagef(t *testing.T) {
	tests := []struct {
		format string
		args   []any
		want   string
	}{
		{format: "", args: nil, want: ""},
		{format: "foo", args: nil, want: "foo"},
		{format: "foo %d", args: []any{42}, want: "foo 42"},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			got1 := Messagef(tt.format, tt.args...)
			got2 := Messagef(tt.format, tt.args...)
			if msg := got1.Error(); msg != tt.want {
				t.Errorf("Messagef(%q, %#v): got: %q, want %q", tt.format, tt.args, msg, tt.want)
			}
			if len(StackTrace(got1)) != 0 {
				t.Errorf("Messagef(%q, %#v): returned error must not contain a stack trace", tt.format, tt.args)
			}
			if got1 == got2 {
				t.Errorf("Messagef(%q, %#v): returned error must not be the same instance", tt.format, tt.args)
			}
		})
	}
}

func TestNew(t *testing.T) {
	// Since New is mostly a wrapper around Join, we only test
	// the error message and stack trace.
	tests := []struct {
		vals    []any
		want    string
		wantNil bool
	}{
		{vals: []any{"foo", "bar"}, want: "foo: bar"},
		{vals: []any{nil}, wantNil: true},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			got := New(tt.vals...)
			switch {
			case tt.wantNil:
				if got != nil {
					t.Errorf("New(%#v): expected nil", tt.vals)
				}
			default:
				if got.Error() != tt.want {
					t.Errorf("New(%#v): got: %q, want %q", tt.vals, got, tt.want)
				}
				st := StackTrace(got)
				if len(st) == 0 {
					t.Errorf("New(%#v): returned error must contain a stack trace", tt.vals)
					return
				}
				if !strings.Contains(st.Frames()[0].Function, "TestNew") {
					t.Errorf("New(%#v): first frame must point to TestNew", tt.vals)
				}
			}
		})
	}
}

func TestNewf(t *testing.T) {
	// Since Newf is mostly a wrapper around Joinf, we only test
	// the error message and stack trace.
	tests := []struct {
		format string
		args   []any
		want   string
	}{
		{format: "foo", args: nil, want: "foo"},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			got := Newf(tt.format, tt.args...)
			if got.Error() != tt.want {
				t.Errorf("Newf(%q, %#v): got: %q, want %q", tt.format, tt.args, got, tt.want)
			}
			st := StackTrace(got)
			if len(st) == 0 {
				t.Errorf("Newf(%q, %#v): returned error must contain a stack trace", tt.format, tt.args)
				return
			}
			if !strings.Contains(st.Frames()[0].Function, "TestNewf") {
				t.Errorf("Newf(%q, %#v): first frame must point to TestNewf", tt.format, tt.args)
			}
		})
	}
}

func TestJoin(t *testing.T) {
	tests := []struct {
		vals    []any
		want    string
		wantNil bool
	}{
		// String
		{vals: []any{""}, want: ""},
		{vals: []any{"foo", "bar"}, want: "foo: bar"},

		// Error
		{vals: []any{Message("foo"), Message("bar")}, want: "foo: bar"},

		// Stringer
		{vals: []any{stringer{s: "foo"}, stringer{s: "bar"}}, want: "foo: bar"},

		// Sprintf
		{vals: []any{42, 314}, want: "42: 314"},

		// Nil cases
		{vals: []any{}, wantNil: true},
		{vals: []any{nil}, wantNil: true},
		{vals: []any{nil, nil}, wantNil: true},
		{vals: []any{nil, "foo", "bar"}, want: "foo: bar"},
		{vals: []any{"foo", nil, "bar"}, want: "foo: bar"},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			got := Join(tt.vals...)
			switch {
			case tt.wantNil:
				if got != nil {
					t.Errorf("Join(%#v): expected nil", tt.vals)
				}
			default:
				if got.Error() != tt.want {
					t.Errorf("Join(%#v): got: %q, want %q", tt.vals, got, tt.want)
				}
				if len(StackTrace(got)) != 0 {
					t.Errorf("Join(%#v): returned error must not contain a stack trace", tt.vals)
				}
				for _, v := range tt.vals {
					if err, ok := v.(error); ok {
						if !errors.Is(got, err) {
							t.Errorf("errors.Is(Join(errs...), err): must return true")
						}
					}
				}
			}
		})
	}
}

func TestJoin_Unwrap(t *testing.T) {
	err1 := Message("first error")
	err2 := Message("second error")
	got := Join(err1, err2)
	unwrapper, ok := got.(interface{ Unwrap() error })
	if !ok {
		t.Fatalf("Join(err1, err2) must implement Unwrap()")
	}
	unwrapped := unwrapper.Unwrap()
	if unwrapped == nil {
		t.Fatalf("Join(err1, err2).Unwrap() must not return nil")
	}
	if !(!errors.Is(unwrapped, err1) && errors.Is(unwrapped, err2)) {
		t.Fatalf("Join(err1, err2).Unwrap() must return the second error")
	}
}

func TestJoinf(t *testing.T) {
	err1 := Message("first error")
	err2 := Message("second error")
	tests := []struct {
		format string
		args   []any
		want   string
	}{
		{format: "simple error", args: nil, want: "simple error"},
		{format: "error with value %d", args: []any{42}, want: "error with value 42"},
		{format: "wrapped error: %w", args: []any{err1}, want: "wrapped error: first error"},
		{format: "multiple errors: %w: %w", args: []any{err1, err2}, want: "multiple errors: first error: second error"},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			got := Joinf(tt.format, tt.args...)
			if got == nil {
				t.Errorf("Joinf(%q, %#v): expected non-nil error", tt.format, tt.args)
				return
			}
			if got.Error() != tt.want {
				t.Errorf("Joinf(%q, %#v): got: %q, want %q", tt.format, tt.args, got, tt.want)
			}
			if len(StackTrace(got)) != 0 {
				t.Errorf("Joinf(%q, %#v): returned error must not contain a stack trace", tt.format, tt.args)
			}
			for _, v := range tt.args {
				if err, ok := v.(error); ok {
					if !errors.Is(got, err) {
						t.Errorf("errors.Is(Joinf(errs...), err): must return true")
					}
				}
			}
		})
	}
}

func TestJoinf_Unwrap(t *testing.T) {
	err1 := Message("first error")
	err2 := Message("second error")
	got := Joinf("%w: %w", err1, err2)
	unwrapper, ok := got.(interface{ Unwrap() error })
	if !ok {
		t.Fatalf("Join(err1, err2) must implement Unwrap()")
	}
	unwrapped := unwrapper.Unwrap()
	if unwrapped == nil {
		t.Fatalf("Join(err1, err2).Unwrap() must not return nil")
	}
	if !(!errors.Is(unwrapped, err1) && errors.Is(unwrapped, err2)) {
		t.Fatalf("Join(err1, err2).Unwrap() must return the second error")
	}
}
