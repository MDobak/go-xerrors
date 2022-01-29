`go-xerrors` is an idiomatic and lightweight package that provides a set of functions to make working with errors
easier. It adds support for stack traces, multierrors, and simplifies working with wrapped errors and panics.
The `go-xerrors` package is fully compatible with Go errors 1.13, supporting the `errors.As`, `errors.Is`,
and `errors.Unwrap` functions.

Main features:

- Stack traces
- Multierrors
- More flexible error warping
- Simplified panic handling

---

# Installation

`go get -u github.com/mdobak/go-xerrors`

# Usage

## Basic errors and stack traces

The most important function in the package is the `xerrors.New` function. This function creates a new error based on the
given message and records the stack trace at the point it was called.

The simplest use of the `xerrors.New` function is to create a simple string-based error along with a stack trace:

```go
err := xerrors.New("access denided")
```

However, calling the `Error` method on the returned error will only return the string that was passed to
the `xerrors.New` function. To retrieve the stack trace, the `xerrors.StackTrace` function can be used. This method will
return an `xerrors.Callers` object, which can be represented as a string using the `fmt` package or by using
the `String` method.

```go
trace := xerrors.StackTrace(err)
fmt.Print(trace)
```

Output:

```
	at main.TestMain (/home/user/app/main_test.go:10)
	at testing.tRunner (/home/user/go /src/testing/testing.go:1259)
	at runtime.goexit (/home/user/go /src/runtime/asm_arm64.s:1133)
```

Another way to display a stack trace is to use the `xerrors.Print`, `xerrors.Sprint`, or `xerrors.Fprint` methods. These
methods automatically detect whether the specified error contains additional information, such as the stack trace, and
display it along with the error message:

```go
xerrors.Print(err)
```

Output:

```
Error: access denided
	at main.TestMain (/home/user/app/main_test.go:10)
	at testing.tRunner (/home/user/go /src/testing/testing.go:1259)
	at runtime.goexit (/home/user/go /src/runtime/asm_arm64.s:1133)
```

## Error wrapping

The `xerrors.New` function accepts not only strings but also other errors. For example, it can be used to add a stack
trace to sentinel errors. The `xerrors` package provides the `xerrors.Message` function, that creates string-based
sentinel errors, to add a stack trace to them, they need to be passed to the `xerrors.New` function:

```jsx
var ErrAccessDenied = xerrors.Message("access denided")
// ...
err := xerrors.New(ErrAccessDenied)
```

Another way to use the `xerrors.New` function is to wrap errors:

```jsx
err := xerrors.New("unable to open resource", ErrAccessDenied)
err.Error() // unable to open resource: access denided
```

It is also possible to wrap an error in another error:

```go
var ErrAccessDenied = xerrors.Message("access denided")
var ErrResourceOpenFailed = xerrors.Message("unable to open resource")
// ...
err := xerrors.New(ErrResourceOpenFailed, ErrAccessDenied)
err.Error() // unable to open resource: access denided
errors.Is(err, ErrResourceOpenFailed) // true
errors.Is(err, ErrAccessDenied) // true
```

Unlike the standard `fmt.Errorf` function, `xerrors.New` keeps references to both errors so that no information is lost
during wrapping.

## Multierrors

Multierrors allow storing a list of errors in a single error, allowing multiple errors to be returned from a function.
It supports the `errors.Is` and `errors.As` methods. However, the `errors.Unwrap` method is not supported.

To create a new multierror, the function `xerrors.Append` should be used. It works similarly to the append function in
the Go language:

```go
var err error
if len(unsername) == 0 {
	err = xerrors.Append(err, xerrors.New("username cannot be empty")
}
if len(password) < 8 {
	err = xerrors.Append(err, xerrors.New("password is too short")
}
```

The error list can be displayed in several ways. The simplest way is to use the `Error` method, which will display
errors as a long, one-line string:

```go
the following errors occurred: [username cannot be empty, password is too short]
```

Most messages returned by the `Error` method are one-line strings; the `xerrors` package follows this convention.

Another way is to use one of the following functions: `xerrors.Print`, `xerrors.Sprint`, or `xerrors.Fprint`. The
advantage of using these functions is that they will also display additional details, such as stack traces, and the
error message is much easier to read:

```
Error: the following errors occurred: [username cannot be empty, password is too short]
1. Error: username cannot be empty
	at xerrors.TestFprint (/home/user/app/main_test.go:10)
	at testing.tRunner (/home/user/go /src/testing/testing.go:1439)
	at runtime.goexit (/home/user/go /src/runtime/asm_arm64.s:1259)
2. Error: password is too short
	at xerrors.TestFprint (/home/user/app/main_test.go:13)
	at testing.tRunner (/home/user/go /src/testing/testing.go:1439)
	at runtime.goexit (/home/user/go /src/runtime/asm_arm64.s:1259)
```

Finally, multierror implements the `xerrors.MultiError` interface, which provides the `Errors` method that returns a
list of errors.

## Recovered panics

In Go, the values returned by the `recover` built-in do not implement the `error` interface, which can be inconvenient.
For this reason, the package provides two functions to convert recovered panics to errors.

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
defer func () {
	if r := recover(); r != nil {
		err := xerrors.FromRecover(r)
		xerrors.Print(err)
	}
}()
```

## Documentation

This package offers a few additional functions and interfaces that may be useful in some use cases. More information
about them can be found in the documentation:

[https://pkg.go.dev/github.com/mdobak/go-xerrors](https://pkg.go.dev/github.com/mdobak/go-xerrors)

## License

Licensed under MIT License
