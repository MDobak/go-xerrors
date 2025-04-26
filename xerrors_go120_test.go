//go:build go1.20
// +build go1.20

package xerrors

import (
	"errors"
	"fmt"
	"testing"
)

func TestJoinf_Go120(t *testing.T) {
	err1 := Message("first error")
	err2 := Message("second error")
	tests := []struct {
		format string
		args   []any
		want   string
	}{
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
