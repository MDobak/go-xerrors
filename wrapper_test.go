package xerrors

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"testing"
)

func TestWrap(t *testing.T) {
	tests := []struct {
		err     error
		wrapper error
		want    string
		wantNil bool
	}{
		{err: Message("err"), wrapper: Message("wrapper"), want: "wrapper: err"},
		{err: io.EOF, wrapper: Message("wrapper"), want: "wrapper: EOF"},
		{err: nil, wrapper: Message("wrapper"), wantNil: true},
		{err: Message("err"), wrapper: nil, want: "err"},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			got := WithWrapper(tt.wrapper, tt.err)
			switch {
			case tt.wantNil:
				if got != nil {
					t.Errorf("WithWrapper(%#v, %#v): expected nil", tt.wrapper, tt.err)
				}
			default:
				if got.Error() != tt.want {
					t.Errorf("WithWrapper(%#v, %#v): got: %q, want %q", tt.wrapper, tt.err, got, tt.want)
				}
				if len(StackTrace(got)) != 0 {
					t.Errorf("WithWrapper(%#v, %#v): returned error must not contain a stack trace", tt.wrapper, tt.err)
				}
				if !errors.Is(got, tt.err) {
					t.Errorf("WithWrapper(%#v, %#v): errors.Is must return true for err", tt.wrapper, tt.err)
				}
				if tt.wrapper != nil && !errors.Is(got, tt.wrapper) {
					t.Errorf("WithWrapper(%#v, %#v): errors.Is must return true for wrapper", tt.wrapper, tt.err)
				}
				if tt.err != nil && !errors.As(got, reflect.New(reflect.TypeOf(tt.err)).Interface()) {
					t.Errorf("errors.As(WithWrapper(%#v, %#v), err): must return true for the err error type", tt.wrapper, tt.err)
				}
				if tt.wrapper != nil && !errors.As(got, reflect.New(reflect.TypeOf(tt.wrapper)).Interface()) {
					t.Errorf("errors.As(WithWrapper(%#v, %#v), err): must return true for the wrapper error type", tt.wrapper, tt.err)
				}
			}
		})
	}
}
