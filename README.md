# go-xerrors

`go-xerrors` is an idiomatic and lightweight Go package designed to enhance error handling in Go applications. It provides functions and types that simplify common error handling tasks by adding support for stack traces, combining multiple errors, and simplifying working with panics. `go-xerrors` maintains full compatibility with Go's standard error handling features (Go 1.13+), including `errors.As`, `errors.Is`, and `errors.Unwrap`.

**Main Features:**

- **Stack Traces**: Automatically captures and attaches stack traces to errors upon creation, significantly aiding in debugging and pinpointing the origin of issues.
- **Multi-Errors**: Allows for the aggregation of multiple errors into a single error instance, useful for reporting all failures from operations that involve multiple steps or components.
- **Flexible Error Wrapping**: Provides ways to wrap errors with additional context or messages. Supports wrapping multiple underlying errors simultaneously while preserving the ability to inspect each one individually.
- **Simplified Panic Handling**: Offers functions to convert recovered panic values into standard Go errors, complete with stack traces, facilitating more robust error recovery logic.

---

## Installation

```bash
go get -u github.com/mdobak/go-xerrors
```

## Usage

### Basic Errors and Stack Traces

A primary use of `go-xerrors` is creating errors that automatically include a stack trace via the `xerrors.New` function:

```go
err := xerrors.New("something went wrong")
```

Invoking the standard `Error()` method on this `err` returns only the message ("something went wrong"), adhering to the Go convention of providing a concise error description.

To display the error with its associated stack trace and other potential details, use the `xerrors.Print`, `xerrors.Sprint`, or `xerrors.Fprint` functions. These functions are designed to format errors created by `go-xerrors`, including their detailed information.

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

To retrieve only the stack trace information programmatically, use `xerrors.StackTrace`. This function returns an `xerrors.Callers` object, which can be formatted as a string.

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

### Sentinel Errors

Sentinel errors are predefined error values representing specific, known failure conditions (e.g., `io.EOF`). They are typically declared as package-level variables. Using sentinel errors allows for reliable error checking using `errors.Is`, avoiding direct string comparisons of error messages.

`go-xerrors` provides `xerrors.Message` to create distinct sentinel error values with consistent messages.

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

### Error Wrapping

The `xerrors.New` function can wrap existing errors, which is useful for adding stack traces or providing additional contextual information.

**Adding Stack Traces to Existing Errors:**

If you receive an error that lacks a stack trace, you can wrap it using `xerrors.New`:

```go
output, err := json.Marshal(data)
if err != nil {
	return xerrors.New("failed to marshal data", err)
}
```

**Wrapping Errors with Additional Context:**

Provide a descriptive string as the first argument to `xerrors.New`, followed by the error(s) to wrap:

```go
if err := updateUserProfile(user); err != nil {
	return xerrors.New("failed to update user profile", err)
}
```

**Wrapping Multiple Errors:**

`xerrors.New` can wrap multiple errors simultaneously:

```go
var ErrConnectionFailed = xerrors.Message("connection failed")
var ErrTimeout = xerrors.Message("operation timed out")

// Wrap both errors under a single error
combinedErr := xerrors.New(ErrConnectionFailed, ErrTimeout)

fmt.Println(combinedErr.Error()) // connection failed: operation timed out
```

This feature is useful when a high-level operation fails due to multiple underlying issues that need to be reported together.

### Multi-Errors (Error Aggregation)

When performing multiple independent operations where several might fail (e.g., validating multiple inputs, processing batch items), use `xerrors.Append` to collect these individual errors into a single multi-error instance.

`xerrors.Append` behaves like Go's built-in `append` but is specifically designed for aggregating errors.

```go
var err error

if input.Username == "" {
    // Append creates/adds to the multi-error, including stack trace if using xerrors.New
    err = xerrors.Append(err, xerrors.New("username cannot be empty"))
}
if len(input.Password) < 8 {
    err = xerrors.Append(err, xerrors.New("password must be at least 8 characters"))
}

if err != nil {
    fmt.Println(err.Error()) // the following errors occurred: [username cannot be empty, password must be at least 8 characters]

    // Detailed output using xerrors.Print:
    xerrors.Print(err)
    // Output:
    // Error: the following errors occurred: [username cannot be empty, password must be at least 8 characters]
    // 1. Error: username cannot be empty
    // 	at main.validateInput (/path/to/your/file.go:XX)
    // 	... stack trace ...
    // 2. Error: password must be at least 8 characters
    // 	at main.validateInput (/path/to/your/file.go:YY)
    // 	... stack trace ...
}
```

The resulting multi-error implements the standard `error` interface as well as `errors.Is`, `errors.As`, and `errors.Unwrap`, allowing you to check for specific errors or extract them.

**Comparison with Go 1.20 `errors.Join`:**

Go 1.20 introduced `errors.Join` for error aggregation. While serving a similar purpose, `xerrors.Append` (especially when used with errors created by `xerrors.New`) offers some differences:

1.  **Individual Stack Traces**: `go-xerrors` multi-errors preserve the individual stack traces associated with each appended error. `errors.Join` does not inherently manage stack traces for the errors it combines.
2.  **Enhanced Formatting**: `xerrors.Print`, `Sprint`, and `Fprint` provide detailed, structured output for multi-errors, listing each constituent error and its stack trace, which can be beneficial for debugging.
3.  **Consistent `Error()` Output**: The `Error()` method of a `go-xerrors` multi-error consistently produces a concise, single-line summary. The output format of `errors.Join`'s `Error()` method can vary.

### Simplified Panic Handling

Go's `recover()` built-in returns the value passed to `panic`, which has type `any` and doesn't directly implement the `error` interface. `go-xerrors` provides utilities to convert these recovered values into proper errors with stack traces.

**Using `xerrors.Recover`:**

This function wraps `recover()` and executes a callback function only if a panic occurred. The callback receives the panic value converted to an `error` (with stack trace). Use it directly with `defer`.

```go
func handleTask() (err error) {
	defer xerrors.Recover(func(err error) {
		err = xerrors.FromRecover(r) // Convert recovered value to error with stack trace
		log.Printf("Recovered from panic during task handling: %s", xerrors.Sprint(err))
	})

	// ... potentially panicking code ...
	panic("task failed")

	return nil
}
```

**Using `xerrors.FromRecover`:**

If you prefer the standard `recover()` pattern, use `xerrors.FromRecover` to manually convert the recovered value after checking if `recover()` returned non-nil.

```go
func handleTask() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err := xerrors.FromRecover(r) // Convert recovered value to error with stack trace
			log.Printf("Recovered from panic during task handling: %s", xerrors.Sprint(err))
		}
	}()

	// ... potentially panicking code ...
	panic("task failed")

	return nil
}
```

## API Summary

Core components of the `go-xerrors` package:

- `xerrors.New(errors ...any) error`: Creates a new error, automatically capturing a stack trace. Can wrap one or multiple existing errors and optionally add a context message.
- `xerrors.Message(message string) error`: Creates a simple sentinel error value identified by its type and message.
- `xerrors.Append(err error, errs ...error) error`: Aggregates errors. Appends new errors to an existing error (or `nil`), returning a multi-error instance. Preserves stack traces of appended errors.
- `xerrors.Recover(callback func(err error))`: A utility function for use with `defer`. It recovers from panics and invokes the provided callback with the panic value converted to an error (including stack trace).
- `xerrors.FromRecover(recoveredValue any) error`: Converts a value returned by the built-in `recover()` function into an error instance with a stack trace.
- `xerrors.Print(err error)`, `xerrors.Sprint(err error) string`, `xerrors.Fprint(w io.Writer, err error)`: Formatting functions that output errors with detailed information, including stack traces and structured multi-error breakdowns.
- `xerrors.StackTrace(err error) Callers`: Extracts the stack trace information from an error created or wrapped by `go-xerrors`.
- Standard Go compatibility: Fully supports `errors.Is`, `errors.As`, and `errors.Unwrap` for interoperability.

## Documentation

For comprehensive details on all functions and types, please refer to the full documentation available at:

[https://pkg.go.dev/github.com/mdobak/go-xerrors](https://pkg.go.dev/github.com/mdobak/go-xerrors)

## License

Licensed under the MIT License.
