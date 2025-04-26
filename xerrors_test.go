package xerrors

import (
	"errors"
	"fmt"
	"io"
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
			err := Message(tt.val)
			if got := err.Error(); got != tt.want {
				t.Errorf("Message(%#v): got: %q, want %q", tt.val, got, tt.want)
			}
			if len(StackTrace(err)) != 0 {
				t.Errorf("Message(%#v): returned error must not contain a stack trace", tt.val)
			}
		})
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		vals    []any
		want    string
		wantNil bool
	}{
		{vals: []any{""}, want: ""},
		{vals: []any{"foo", "bar"}, want: "foo: bar"},
		{vals: []any{nil, "foo", "bar"}, want: "foo: bar"},
		{vals: []any{"foo", nil, "bar"}, want: "foo: bar"},
		{vals: []any{Message("foo"), Message("bar")}, want: "foo: bar"},
		{vals: []any{io.EOF, io.EOF}, want: "EOF: EOF"},
		{vals: []any{stringer{s: "foo"}, stringer{s: "bar"}}, want: "foo: bar"},
		{vals: []any{42, 314}, want: "42: 314"},
		{vals: []any{}, wantNil: true},
		{vals: []any{nil}, wantNil: true},
		{vals: []any{nil, nil}, wantNil: true},
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
				if len(StackTrace(got)) == 0 {
					t.Errorf("New(%#v): returned error must contain a stack trace", tt.vals)
				}
				for _, v := range tt.vals {
					if err, ok := v.(error); ok {
						if !errors.Is(got, err) {
							t.Errorf("errors.Is(New(errs...), err): must return true")
						}
					}
				}
			}
		})
	}
}
