# go-xerrors

[![Go Reference](https://pkg.go.dev/badge/github.com/mdobak/go-xerrors.svg)](https://pkg.go.dev/github.com/mdobak/go-xerrors) [![Go Report Card](https://goreportcard.com/badge/github.com/mdobak/go-xerrors)](https://goreportcard.com/report/github.com/mdobak/go-xerrors) [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT) [![Coverage Status](https://coveralls.io/repos/github/MDobak/go-xerrors/badge.svg?branch=coverage)](https://coveralls.io/github/MDobak/go-xerrors?branch=coverage)

`go-xerrors` is a small, idiomatic library that makes error handling in Go easier. It provides utilities for creating errors with stack traces, wrapping existing errors, aggregating multiple errors, and recovering from panics.

**Main features**

- **Stack traces**: Capture stack traces when creating errors to pinpoint the source during debugging.
- **Multi-errors**: Aggregate multiple errors into a single error while preserving individual context.
- **Panic handling**: Convert panic values to standard Go errors with stack traces.
- **Zero dependencies**: No external dependencies beyond the Go standard library.

> Note: This package is stable. Updates are rare and limited to bug fixes and support for new Go versions or error-related features. Since 1.0, the API has been frozen. No breaking changes will be introduced in future releases.

---

## Installation

```bash
go get -u github.com/mdobak/go-xerrors
```

## Usage

### Example

Here’s a quick example of creating and handling errors with `go-xerrors`.

```go
package main

import (
    "database/sql"
    "fmt"

    "github.com/mdobak/go-xerrors"
)

func findUserByID(id int) error {
    // Simulate a standard library error.
    err := sql.ErrNoRows

    // Wrap the original error with additional context and capture a stack trace
    // at this point in the call stack.
    return xerrors.Newf("user %d not found: %w", id, err)
}

func main() {
    err := findUserByID(123)
    if err != nil {
        // 1) err.Error() returns a concise, log-friendly message.
        fmt.Println(err.Error())
        // Output:
        // user 123 not found: sql: no rows in result set

        // 2) xerrors.Print shows a detailed message with a stack trace.
        xerrors.Print(err)
        // Output:
        // Error: user 123 not found: sql: no rows in result set
        //     at main.findUserByID (/home/user/app/main.go:15)
        //     at main.main (/home/user/app/main.go:20)
        //     at runtime.main (/usr/local/go/src/runtime/proc.go:250)
        //     at runtime.goexit (/usr/local/go/src/runtime/asm_amd64.s:1594)
    }
}
```

### Creating Errors with Stack Traces

The primary way to create an error in `go-xerrors` is by using the `xerrors.New` or `xerrors.Newf` functions:

```go
// Create a new error with a stack trace.
err := xerrors.New("something went wrong")

// Create a formatted error with a stack trace.
err := xerrors.Newf("something went wrong: %s", reason)
```

Calling `err.Error()` returns only the message ("something went wrong"), following the Go convention of keeping error strings concise.

### Displaying Detailed Errors

To print the error with its stack trace and details, use `xerrors.Print`, `xerrors.Sprint`, or `xerrors.Fprint`: 

```go
xerrors.Print(err)
```

Output:

```
Error: something went wrong
	at main.main (/home/user/app/main.go:10)
	at runtime.main (/usr/local/go/src/runtime/proc.go:225)
	at runtime.goexit (/usr/local/go/src/runtime/asm_amd64.s:1371)
```

### Working with Stack Traces

To retrieve the stack trace information programmatically:

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

You can also add a stack trace to an existing error with `xerrors.WithStackTrace`, and choose how many frames to skip. This is handy when a helper creates errors but you don’t want its own frame captured:

```go
func errNotFound(path string) error {
	return xerrors.WithStackTrace(ErrNotFound{path: path}, 1) // skip one frame
}
```

### Wrapping Errors

You can also wrap existing errors:

```go
output, err := json.Marshal(data)
if err != nil {
	return xerrors.New("failed to marshal data", err)
}
```

With formatted messages:

```go
output, err := json.Marshal(data)
if err != nil {
	return xerrors.Newf("failed to marshal data %v: %w", data, err)
}
```

> Wrapping multiple errors with `xerrors.Newf` is possible only in Go 1.20 and later.

### Creating Error Chains Without Stack Traces

When you don’t need a stack trace (for example, when creating sentinel errors), use `xerrors.Join` and `xerrors.Joinf`: 

```go
err := xerrors.Join("operation failed", otherErr)
```

With formatted messages:

```go
err := xerrors.Joinf("operation failed: %w", otherErr)
```

> Wrapping multiple errors with `xerrors.Joinf` is possible only in Go 1.20 and later.

The main difference between Go's `fmt.Errorf` and `xerrors.Newf`/`xerrors.Joinf` is that the latter functions preserve the error chain, whereas `fmt.Errorf` flattens it (i.e., its `Unwrap` method returns all underlying errors at once instead of just the next one in the chain).

### Sentinel Errors

Sentinel errors are predefined, exported error values used to signal specific, well-known conditions (e.g., `io.EOF`). The `go-xerrors` package provides the `xerrors.Message` and `xerrors.Messagef` functions to create distinct sentinel error values:

```go
var ErrAccessDenied = xerrors.Message("access denied")

// ...

func performAction() error {
	// ...
	return ErrAccessDenied
}

// ...

err := performAction()
if errors.Is(err, ErrAccessDenied) {
    log.Println("Operation failed due to access denial.")
}
```

For formatted sentinel errors:

```go
const MaxLength = 10
var ErrInvalidInput = xerrors.Messagef("max length of %d exceeded", MaxLength)
```

### Multi-Errors

When performing multiple independent operations where several might fail, use `xerrors.Append` to collect these individual errors into a single multi-error instance:

```go
var err error

if input.Username == "" {
    err = xerrors.Append(err, xerrors.New("username cannot be empty"))
}
if len(input.Password) < 8 {
    err = xerrors.Append(err, xerrors.New("password must be at least 8 characters"))
}

if err != nil {
    fmt.Println(err.Error()) 
	// Output:
	// [username cannot be empty, password must be at least 8 characters]

    // Detailed output using xerrors.Print:
    xerrors.Print(err)
    // Output:
    // Error: [username cannot be empty, password must be at least 8 characters]
    // 1. Error: username cannot be empty
    //     at main.validateInput (/home/user/app/main.go:40)
    //     at main.main (/home/user/app/main.go:20)
    //     at runtime.main (/usr/local/go/src/runtime/proc.go:250)
    //     at runtime.goexit (/usr/local/go/src/runtime/asm_amd64.s:1594)
    // 2. Error: password must be at least 8 characters
    //     at main.validateInput (/home/user/app/main.go:43)
    //     at main.main (/home/user/app/main.go:20)
    //     at runtime.main (/usr/local/go/src/runtime/proc.go:250)
    //     at runtime.goexit (/usr/local/go/src/runtime/asm_amd64.s:1594)
}
```

The resulting multi-error implements the standard `error` interface as well as `errors.Is`, `errors.As`, and `errors.Unwrap`, allowing you to check for specific errors or extract them.

**Comparison with Go 1.20 `errors.Join`:**

Go 1.20 introduced `errors.Join` for error aggregation. While it serves a similar purpose, `xerrors.Append` preserves the individual stack traces associated with each appended error and adheres to the convention of returning a single line from the `Error()` method.

### Simplified Panic Handling

Panics can be challenging to locate and handle effectively in Go applications, especially when using `recover()`. Common issues, such as nil pointer dereferences or out-of-bounds slice accesses, often result in unclear panic messages. Without a stack trace, pinpointing the origin of the panic can be difficult.

`go-xerrors` provides utilities to convert panic values into proper errors with stack traces.

**Using `xerrors.Recover`:**

```go
func handleTask() (err error) {
	defer xerrors.Recover(func(err error) {
		log.Printf("Recovered from panic during task handling: %s", xerrors.Sprint(err))
	})

	// ... potentially panicking code ...

	return nil
}
```

**Using `xerrors.FromRecover`:**

```go
func handleTask() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = xerrors.FromRecover(r) // Convert recovered value to error with stack trace
			log.Printf("Recovered from panic during task handling: %s", xerrors.Sprint(err))
		}
	}()

	// ... potentially panicking code ...

	return nil
}
```

The returned error implements the `PanicError` interface, which provides access to the original panic value via the `Panic()` method.

### Choosing Between `New`, `Join`, and `Append`

While these functions can all be used to aggregate errors, they each serve distinct purposes:

- **`xerrors.New`**: Use this to create errors and attach stack traces, especially when wrapping existing errors to provide additional context.
- **`xerrors.Join`**: Use this to chain errors together _without_ capturing stack traces.
- **`xerrors.Append`**: Use this to aggregate multiple, independent errors into a single multi-error. This is useful when several operations might fail, and you want to report all failures at once.

#### Examples

##### Error with Stack Trace

```go
func (m *MyStruct) MarshalJSON() ([]byte, error) {
	output, err := json.Marshal(m)
	if err != nil {
		// Wrap the error with additional context and capture a stack trace.
		return nil, xerrors.New("failed to marshal data", err)
	}
	return output, nil
}
```

##### Sentinel Errors

```go
var (
	// Using xerrors.Join allows us to create sentinel errors that can be
	// checked with errors.Is against both ErrValidation and the
	// specific validation error. We do not want to capture a stack trace
	// here; therefore, we use xerrors.Join instead of xerrors.New.
	ErrValidation   = xerrors.Message("validation error")
	ErrInvalidName  = xerrors.Join(ErrValidation, "name is invalid")
	ErrInvalidAge   = xerrors.Join(ErrValidation, "age is invalid")
	ErrInvalidEmail = xerrors.Join(ErrValidation, "email is invalid")
)

func (m *MyStruct) Validate() error {
	if !m.isNameValid() {
		return xerrors.New(ErrInvalidName)
	}
	if !m.isAgeValid() {
		return xerrors.New(ErrInvalidAge)
	}
	if !m.isEmailValid() {
		return xerrors.New(ErrInvalidEmail)
	}
	return nil
}
```

##### Multi-Error Validation

```go
func (m *MyStruct) Validate() error {
	var err error
	if m.Name == "" {
		err = xerrors.Append(err, xerrors.New("name cannot be empty"))
	}
	if m.Age < 0 {
		err = xerrors.Append(err, xerrors.New("age cannot be negative"))
	}
	if m.Email == "" {
		err = xerrors.Append(err, xerrors.New("email cannot be empty"))
	}
	return err
}
```

## API Reference

### Core Functions

- `xerrors.New(errors ...any) error`: Creates a new error with a stack trace
- `xerrors.Newf(format string, args ...any) error`: Creates a formatted error with a stack trace
- `xerrors.Join(errors ...any) error`: Creates a chained error without a stack trace
- `xerrors.Joinf(format string, args ...any) error`: Creates a formatted chained error without a stack trace
- `xerrors.Message(message string) error`: Creates a simple sentinel error
- `xerrors.Messagef(format string, args ...any) error`: Creates a simple formatted sentinel error
- `xerrors.Append(err error, errs ...error) error`: Aggregates errors into a multi-error

### Panics

- `xerrors.Recover(callback func(err error))`: Recovers from panics and invokes a callback with the error
- `xerrors.FromRecover(recoveredValue any) error`: Converts a recovered value to an error with a stack trace

### Stack Trace

- `xerrors.StackTrace(err error) Callers`: Extracts the stack trace from an error
- `xerrors.WithStackTrace(err error, skip int) error`: Wraps an error with a stack trace
- `DefaultCallersFormatter`: The default formatter for Callers, used when printing stack traces.
- `DefaultFrameFormatter`: The default formatter for Frame, used when printing stack traces.

### Error printing

- `xerrors.Print(err error)`: Prints a formatted error to stderr
- `xerrors.Sprint(err error) string`: Returns a formatted error as a string
- `xerrors.Fprint(w io.Writer, err error)`: Writes a formatted error to the provided writer

### Interfaces

- `DetailedError`: For errors that provide detailed information
- `PanicError`: For errors created from panic values with access to the original panic value

## Documentation

For full API details, see the documentation:

[https://pkg.go.dev/github.com/mdobak/go-xerrors](https://pkg.go.dev/github.com/mdobak/go-xerrors)

## License

Licensed under the MIT License.
