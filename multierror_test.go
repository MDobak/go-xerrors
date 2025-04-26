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
		{err: nil, errs: []error{Message("a"), Message("b")}, want: "the following errors occurred: [a, b]"},
		{err: nil, errs: []error{nil, Message("a")}, want: "a"},
		{err: Message("a"), errs: []error{Message("b"), Message("c")}, want: "the following errors occurred: [a, b, c]"},
		{err: Message("a"), errs: nil, want: "a"},
		{err: multiError{Message("a"), Message("b")}, errs: nil, want: "the following errors occurred: [a, b]"},
		{err: multiError{Message("a"), Message("b")}, errs: []error{Message("c")}, want: "the following errors occurred: [a, b, c]"},
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

func TestMultiError_ErrorDetails(t *testing.T) {
	tests := []struct {
		errs   []error
		want   string
		regexp bool
	}{
		{errs: []error{}, want: ``},
		{errs: []error{Message("a")}, want: "1. Error: a\n"},
		{errs: []error{Message("a"), Message("b")}, want: "1. Error: a\n2. Error: b\n"},
		{errs: []error{Message("a"), multiError{Message("b"), Message("c")}}, want: "1. Error: a\n2. Error: the following errors occurred: [b, c]\n\t1. Error: b\n\t2. Error: c\n"},
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
