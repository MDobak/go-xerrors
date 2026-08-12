package xerrors

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
)

func TestAppend(t *testing.T) {
	tests := []struct {
		err     error
		errs    []error
		want    string
		wantNil bool
	}{
		{err: nil, errs: []error{Message("a"), Message("b")}, want: "[a, b]"},
		{err: nil, errs: []error{nil, Message("a")}, want: "a"},
		{err: Message("a"), errs: []error{Message("b"), Message("c")}, want: "[a, b, c]"},
		{err: Message("a"), errs: nil, want: "a"},
		{err: multiError{Message("a"), Message("b")}, errs: nil, want: "[a, b]"},
		{err: multiError{Message("a"), Message("b")}, errs: []error{Message("c")}, want: "[a, b, c]"},
		{err: multiError{}, errs: []error{Message("a"), nil}, want: "a"},
		{err: nil, errs: nil, wantNil: true},
		{err: nil, errs: []error{nil, nil}, wantNil: true},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			got := Append(tt.err, tt.errs...)
			if tt.wantNil {
				if got != nil {
					t.Errorf("Append(nil): must return nil")
				}
			} else {
				if got.Error() != tt.want {
					t.Errorf("Append(err, errs...).Error(): got: %q, want %q", got, tt.want)
				}
				if len(StackTrace(got)) != 0 {
					t.Errorf("Append(err, errs...): returned error must not contain a stack trace")
				}
				if errors.Is(got, Message("foo")) {
					t.Errorf("errors.Is(Append(err, errs...), err): must return false for not included error")
				}
				if errors.As(got, reflect.New(reflect.TypeOf(&withWrapper{})).Interface()) {
					t.Errorf("errors.As(Append(err, errs...), err): must return false for a different error type")
				}
				for _, err := range tt.errs {
					if err == nil {
						continue
					}
					if !errors.Is(got, err) {
						t.Errorf("errors.Is(Append(err, errs...), errs[n]): must return true for all errors")
					}
					if !errors.As(got, reflect.New(reflect.TypeOf(err)).Interface()) {
						t.Errorf("errors.As(Append(err, errs...), errs[n]): must return true for all errors")
					}
				}
			}
		})
	}
}

func TestAppend_DoesNotModifyArgument(t *testing.T) {
	// Appending to the same error twice must return two independent
	// errors. The second call must not overwrite the error appended by
	// the first one.
	base := Append(nil, Message("a"), Message("b"), Message("c"))
	first := Append(base, Message("d"))
	second := Append(base, Message("e"))
	if got, want := base.Error(), "[a, b, c]"; got != want {
		t.Errorf("Append(err, errs...): must not modify the error passed as an argument, got: %q, want: %q", got, want)
	}
	if got, want := first.Error(), "[a, b, c, d]"; got != want {
		t.Errorf("Append(err, errs...): got: %q, want: %q", got, want)
	}
	if got, want := second.Error(), "[a, b, c, e]"; got != want {
		t.Errorf("Append(err, errs...): got: %q, want: %q", got, want)
	}
}

func TestMultiError_ErrorDetails(t *testing.T) {
	tests := []struct {
		errs   []error
		want   string
		regexp bool
	}{
		{errs: []error{}, want: ``},
		{errs: []error{Message("a")}, want: "1. Error: a\n"},
		{errs: []error{Message("a"), Message("b")}, want: "1. Error: a\n2. Error: b\n"},
		{errs: []error{Message("a"), multiError{Message("b"), Message("c")}}, want: "1. Error: a\n2. Error: [b, c]\n\t1. Error: b\n\t2. Error: c\n"},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			err := multiError(tt.errs)
			if got := err.ErrorDetails(); got != tt.want {
				t.Errorf("multiError(errs).ErrorDetails(): %q does not match %q", got, tt.want)
			}
		})
	}
}
