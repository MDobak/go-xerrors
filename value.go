package xerrors

import (
	"errors"
	"fmt"
	"reflect"
)

// WithValue adds a value to an error.
func WithValue(err error, key string, val interface{}) error {
	if err == nil {
		return nil
	}
	return &value{
		err:   err,
		key:   key,
		value: val,
	}
}

type value struct {
	err   error
	key   string
	value interface{}
}

// Format implements the fmt.Formatter interface.
//
// The verbs:
//
//	%s	an error
//	%v	same as %s, the plus or hash flags print the value associated with the error
func (err *value) Format(s fmt.State, verb rune) {
	format(s, verb, err.err)
	if verb == 'v' && (s.Flag('+') || s.Flag('#')) {
		typeOf := reflect.TypeOf(err.value)
		of := reflect.ValueOf(err.value)
		switch typeOf.Kind() {
		case reflect.Slice, reflect.Array, reflect.Chan, reflect.Map, reflect.String, reflect.Ptr:
			_, _ = fmt.Fprintf(s, "\nvalue %q = (%s) (len=%d) \"%v\"", err.key, typeOf, of.Len(), of)
		default:
			_, _ = fmt.Fprintf(s, "\nvalue %q = (%s) \"%v\"", err.key, typeOf, of)
		}
	}
}

func (err *value) Value() (key string, value interface{}) {
	return err.key, err.value
}

func (err *value) Error() string { return err.err.Error() }
func (err *value) Unwrap() error { return err.err }

// Values returns the values associated to an error.
func Values(err error) map[string]interface{} {
	vals := make(map[string]interface{})
	for ; err != nil; err = errors.Unwrap(err) {
		err, ok := err.(*value)
		if !ok {
			continue
		}
		k, v := err.Value()
		_, ok = vals[k]
		if ok {
			continue
		}
		vals[k] = v
	}
	return vals
}
