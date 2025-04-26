package xerrors

import (
	"fmt"
	"strings"
	"testing"
)

type testErr struct {
	err     string
	details string
	wrapped error
}

func (e testErr) Error() string {
	return e.err
}

func (e testErr) DetailedError() string {
	return e.err + "\n" + e.details + "\n"
}

func (e testErr) Unwrap() error {
	return e.wrapped
}

func TestFormat(t *testing.T) {
	tests := []struct {
		err  error
		want string
	}{
		{
			err: Message("foo"), want: "Error: foo\n",
		},
		{
			err:  testErr{err: "err", details: "details"},
			want: "Error: err\ndetails\n",
		},
		{
			err:  testErr{err: "err", details: "details", wrapped: Message("wrapped")},
			want: "Error: err\ndetails\n",
		},
		{
			err:  testErr{err: "err", details: "details", wrapped: testErr{err: "wrapped err", details: "wrapped details"}},
			want: "Error: err\ndetails\nPrevious error: wrapped err\nwrapped details\n",
		},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			if got := Sprint(tt.err); got != tt.want {
				t.Errorf("Sprint(%#v): %q does not match %q", tt.err, got, tt.want)
			}
		})
	}
}

func TestPrint(t *testing.T) {
	prevErrWriter := errWriter
	defer func() { errWriter = prevErrWriter }()

	err := Message("foo")
	buf := &strings.Builder{}
	errWriter = buf
	Print(err)
	got := buf.String()
	exp := "Error: foo\n"
	if got != exp {
		t.Errorf("Print(buf, %#v): wrote invalid error message, got %q but %q expected", err, got, exp)
	}
}

func TestSprint(t *testing.T) {
	a := New("access denied")
	Print(a)

	err := Message("foo")
	got := Sprint(err)
	exp := "Error: foo\n"
	if got != exp {
		t.Errorf("Sprint(b, %#v): wrote invalid error message, got %q but %q expected", err, got, exp)
	}
}

func TestFprint(t *testing.T) {
	err := Message("foo")
	buf := &strings.Builder{}
	n, werr := Fprint(buf, err)
	got := buf.String()
	exp := "Error: foo\n"
	if werr != nil {
		t.Errorf("Fprint(buf, %#v): returned an error", err)
	}
	if n != 11 {
		t.Errorf("Fprint(buf, %#v): returned invalid number of bytes", err)
	}
	if got != exp {
		t.Errorf("Fprint(buf, %#v): wrote invalid error message, got %q but %q expected", err, got, exp)
	}
}
