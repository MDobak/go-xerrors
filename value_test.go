package xerrors

import (
	"fmt"
	"reflect"
	"testing"
)

func TestValue(t *testing.T) {
	err := New("error")
	err = WithValue(err, "foo", "bar")
	vals := Values(err)
	want := map[string]interface{}{
		"foo": "bar",
	}
	if !reflect.DeepEqual(vals, want) {
		t.Errorf("Values() = %v, want %v", vals, want)
	}
}

func TestValueOverWrite(t *testing.T) {
	err := New("error")
	err = WithValue(err, "test", 1)
	err = WithValue(err, "test", 2)
	vals := Values(err)
	want := map[string]interface{}{
		"test": 2,
	}
	if !reflect.DeepEqual(vals, want) {
		t.Errorf("Values() = %v, want %v", vals, want)
	}
}

func TestValueNil(t *testing.T) {
	err := WithValue(nil, "foo", "bar")
	if err != nil {
		t.Fatal(err)
	}
}

func TestValuesEmpty(t *testing.T) {
	err := New("error")
	vals := Values(err)
	if len(vals) != 0 {
		t.Fatalf("values not empty: got %#v", vals)
	}
}

func TestValueError(t *testing.T) {
	err := New("error")
	err = WithValue(err, "foo", "bar")
	s := fmt.Sprint(err)
	want := "error"
	if s != want {
		t.Fatalf("unexpected message: got %q, want %q", s, want)
	}
}

func TestValueFormat(t *testing.T) {
	tests := []struct {
		format string
		value  interface{}
		want   string
	}{
		{format: "%s", value: "bar", want: "error"},
		{format: "%+v", value: "bar", want: "error\nvalue \"foo\" = (string) (len=3) \"bar\""},
		{format: "%+v", value: 4, want: "error\nvalue \"foo\" = (int) \"4\""},
	}
	for n, tt := range tests {
		t.Run(fmt.Sprintf("case-%d", n+1), func(t *testing.T) {
			err := New("error")
			err = WithValue(err, "foo", tt.value)
			s := fmt.Sprintf(tt.format, err)
			if s != tt.want {
				t.Fatalf("unexpected message: got %q, want %q", s, tt.want)
			}
		})
	}
}
