# go-xerrors

`go-xerrors` is an idiomatic and lightweight package that provides a set of functions to make working with errors
easier. It adds support for stack traces, multierrors, and simplifies working with wrapped errors and panics.
The `go-xerrors` package is fully compatible with Go errors 1.13, supporting the `errors.As`, `errors.Is`,
and `errors.Unwrap` functions.

**Main features:**

- Stack traces
- Multierrors
- More flexible error warping
- Simplified panic handling

---

## Installation

`go get -u github.com/mdobak/go-xerrors`

## Usage

### Basic errors and stack traces

The most basic usage of `go-xerrors` is to create a new error with a stack trace. This can be done with the
`xerrors.New` function. The simplest usage is to pass a string, which will be used as the error message.

```go
err := xerrors.New("something went wrong")
```

However, calling the `Error` method on this error will only return the error message, not the stack trace. To
get the stack trace, the `xerrors.StackTrace` function can be used. This function will return an `xerrors.Callers`
object, which contains the stack trace. The `String` method on this object can be used to get a string representation
of the stack trace.

```go
trace := xerrors.StackTrace(err)
fmt.Print(trace)
```

Output:

```
	at main.TestMain (/home/user/app/main_test.go:10)
	at testing.tRunner (/home/user/go/src/testing/testing.go:1259)
	at runtime.goexit (/home/user/go/src/runtime/asm_arm64.s:1133)
```

Another way to display the stack trace is to use the `xerrors.Print`, `xerrors.Sprint`, or `xerrors.Fprint` functions.
These functions will detect if the error passed contains a stack trace and print it to the stderr if it does. It is done
by checking if the error implements the `xerrors.DetailedError` interface. This interface has a single method,
`ErrorDetails`, that returns an additional information about the error, such as the stack trace.

```go
xerrors.Print(err)
```

Output:

```
Error: access denied
	at main.TestMain (/home/user/app/main_test.go:10)
	at testing.tRunner (/home/user/go/src/testing/testing.go:1259)
	at runtime.goexit (/home/user/go/src/runtime/asm_arm64.s:1133)
```

The reason why standard `Error()` method does not return the stack trace is because most developers expect the `Error()`
method to return only a one-line error message without punctuation at the end. This library follows this convention.

### Sentinel errors

Sentinel errors are errors that are defined as constants. They are useful to check if an error is of a specific type
without having to compare the error message. This library provides a `xerrors.Message` function that can be used to
create a sentinel error.

```go
var ErrAccessDenied = xerrors.Message("access denied")
// ...
if errors.Is(err, ErrAccessDenied) {
    // ...
}
```

### Error wrapping

The `xerrors.New` function accepts not only strings but also other errors. For example, it can be used to add a stack
trace to sentinel errors.

```go
var ErrAccessDenied = xerrors.Message("access denied")
// ...
err := xerrors.New(ErrAccessDenied)
//
if errors.Is(err, ErrAccessDenied) {
    xerrors.Print(err) // prints error along with the stack trace
}
```

Another way to use the `xerrors.New` function is to wrap an existing error with a new error message.

```go
err := xerrors.New("unable to open resource", ErrAccessDenied)
fmt.Print(err.Error()) // unable to open resource: access denied
```

It is also possible to wrap an error with another error. Unlike the `fmt.Errorf` function, references to both errors 
will be preserved, so it is possible to check if the new error is one of the wrapped errors.

```go
var ErrAccessDenied = xerrors.Message("access denied")
var ErrResourceOpenFailed = xerrors.Message("unable to open resource")
// ...
err := xerrors.New(ErrResourceOpenFailed, ErrAccessDenied)
fmt.Print(err.Error()) // unable to open resource: access denied
errors.Is(err, ErrResourceOpenFailed) // true
errors.Is(err, ErrAccessDenied) // true
```

### Multierrors

Multierrors are a set of errors that can be treated as a single error. The `xerrors` package provides the 
`xerrors.Append` function to create them. It works similarly to the append function in the Go language. The function 
accepts a variadic number of errors and returns a new error that contains all of them. The returned error supports 
`errors.Is` and `errors.As` methods. However, the `errors.Unwrap`method is not supported.

```go
var err error
if len(unsername) == 0 {
    err = xerrors.Append(err, xerrors.New("username cannot be empty"))
}
if len(password) < 8 {
    err = xerrors.Append(err, xerrors.New("password is too short"))
}
```

The error list can be displayed in several ways. The simplest way is to use the `Error` method, which will display
errors as a long, one-line string:

```
the following errors occurred: [username cannot be empty, password is too short]
```

Another way is to use one of the following functions: `xerrors.Print`, `xerrors.Sprint`, or `xerrors.Fprint`. The
advantage of using these functions is that they will also print additional details, such as stack traces, and the
error message is much easier to read:

```
Error: the following errors occurred: [username cannot be empty, password is too short]
1. Error: username cannot be empty
	at xerrors.TestFprint (/home/user/app/main_test.go:10)
	at testing.tRunner (/home/user/go/src/testing/testing.go:1439)
	at runtime.goexit (/home/user/go/src/runtime/asm_arm64.s:1259)
2. Error: password is too short
	at xerrors.TestFprint (/home/user/app/main_test.go:13)
	at testing.tRunner (/home/user/go/src/testing/testing.go:1439)
	at runtime.goexit (/home/user/go/src/runtime/asm_arm64.s:1259)
```

Finally, multierror implements the `xerrors.MultiError` interface, which provides the `Errors` method that returns a
list of errors.

### Recovered panics

In Go, the values returned by the `recover` built-in do not implement the `error` interface, which may be inconvenient.
This library provides two functions to easily convert a recovered value into an error.

The first function, `xerrors.Recover`, works similarly to the `recover` built-in. This function must always be called
directly using the `defer` keyword. The callback will only be called during a panic, and the provided error will contain
a stack trace:

```go
defer xerrors.Recover(func (err error) {
    xerrors.Print(err)
})
```

The second function allows converting a value returned from `recover` built-in to an error with a stack trace:

```go
defer func() {
    if r := recover(); r != nil {
        err := xerrors.FromRecover(r)
        xerrors.Print(err)
    }
}()
```

### Documentation

This package offers a few additional functions and interfaces that may be useful in some use cases. More information
about them can be found in the documentation:

[https://pkg.go.dev/github.com/mdobak/go-xerrors](https://pkg.go.dev/github.com/mdobak/go-xerrors)

### License

Licensed under MIT License
