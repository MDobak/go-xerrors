# go-xerrors

`go-xerrors` is an idiomatic, lightweight Go package designed to enhance error handling in Go applications. It provides functions and types that simplify common error handling tasks by adding support for stack traces, combining multiple errors, and simplifying work with panics. `go-xerrors` maintains full compatibility with Go's standard error handling features (including changes in Go 1.13 and 1.20), such as `errors.As`, `errors.Is`, and `errors.Unwrap`.

**Main Features:**

- **Stack Traces**: Automatically captures and attaches stack traces to errors upon creation, which significantly aids debugging and helps pinpoint the origin of issues.
- **Multi-Errors**: Enables the aggregation of multiple errors into a single error instance, useful for reporting all failures from operations that involve multiple steps or components.
- **Flexible Error Wrapping**: Provides ways to wrap errors with additional context or messages, while preserving the ability to inspect each underlying error individually.
- **Simplified Panic Handling**: Provides functions for converting recovered panic values into standard Go errors with stack traces, facilitating more robust error recovery logic.

---

## Installation

```bash
go get -u github.com/mdobak/go-xerrors
```

## Basic Usage

### Creating Errors with Stack Traces

The primary way to create an error in `go-xerrors` is by using the `xerrors.New` or `xerrors.Newf` functions:

```go
// Create a new error with a stack trace
err := xerrors.New("something went wrong")

// Create a formatted error with a stack trace
err := xerrors.Newf("something went wrong: %s", reason)
```

Calling the standard `Error()` method on `err` returns only the message ("something went wrong"), adhering to the Go convention of providing a concise error description.

### Displaying Detailed Errors

To display the error with the associated stack trace and additional details, use the `xerrors.Print`, `xerrors.Sprint`, or `xerrors.Fprint` functions:

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

To retrieve only the stack trace information programmatically:

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

You can also explicitly add a stack trace to an existing error:

```go
err := someFunction()
errWithStack := xerrors.WithStackTrace(err, 0) // 0 skips no frames
```

### Wrapping Errors

The `xerrors.New` and `xerrors.Newf` functions can also wrap existing errors while preserving their stack traces:

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

### Creating Error Chains Without Stack Traces

For situations where you don't need a stack trace (such as creating sentinel errors), use `xerrors.Join` and `xerrors.Joinf`:

```go
err := xerrors.Join("operation failed", existingError)
```

The key difference between `fmt.Errorf` and `xerrors.Newf`/`xerrors.Joinf` is that the latter functions preserve the error chain, whereas `fmt.Errorf` flattens it (i.e., its `Unwrap` method returns all underlying errors at once, instead of just the next one in the chain).

### Sentinel Errors

Sentinel errors are predefined error values representing specific, known failure conditions. `go-xerrors` provides `xerrors.Message` to create distinct sentinel error values with consistent messages:

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

Go 1.20 introduced `errors.Join` for error aggregation. While serving a similar purpose, `xerrors.Append` offers:

1. **Individual Stack Traces**: Preserves the individual stack traces associated with each appended error
2. **Enhanced Formatting**: Provides detailed, structured output for multi-errors
3. **Consistent `Error()` Output**: Produces a concise, single-line summary

### Simplified Panic Handling

`go-xerrors` provides utilities to convert panic values into proper errors with stack traces.

**Using `xerrors.Recover`:**

```go
func handleTask() (err error) {
	defer xerrors.Recover(func(err error) {
		log.Printf("Recovered from panic during task handling: %s", xerrors.Sprint(err))
	})

	// ... potentially panicking code ...
	panic("task failed")

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
	panic("task failed")

	return nil
}
```

The returned error implements the `PanicError` interface, which provides access to the original panic value via the `Panic()` method.

## When to use `New`, `Join`, or `Append`

While all three functions can be used to aggregate errors, they serve different purposes:

- **`xerrors.New`**: Create errors with stack traces, useful for wrapping existing errors to add context
- **`xerrors.Join`**: Create chained errors without stack traces, useful for defining sentinel errors
- **`xerrors.Append`**: Create multi-errors by aggregating independent errors

## Examples

### Error with Stack Trace

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

### Sentinel Errors

```go
var (
	// Using xerrors.Join lets us create sentinel errors that can be
	// checked with errors.Is against both ErrValidation and the
	// specific validation error. We do not want to capture a stack trace
	// here; hence, we use xerrors.Join instead of xerrors.New.
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

### Multi-Error Validation

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
- `xerrors.WithStackTrace(err error, skip int) error`: Wraps an error with a stack trace

### Multi-Error Functions

- `xerrors.Append(err error, errs ...error) error`: Aggregates errors into a multi-error

### Panic Handling

- `xerrors.Recover(callback func(err error))`: Recovers from panics and invokes a callback with the error
- `xerrors.FromRecover(recoveredValue any) error`: Converts a recovered value to an error with a stack trace

### Formatting and Stack Trace Functions

- `xerrors.Print(err error)`: Prints a formatted error to stderr
- `xerrors.Sprint(err error) string`: Returns a formatted error as a string
- `xerrors.Fprint(w io.Writer, err error)`: Writes a formatted error to the provided writer
- `xerrors.StackTrace(err error) Callers`: Extracts the stack trace from an error

### Key Interfaces

- `DetailedError`: For errors that provide detailed information
- `PanicError`: For errors created from panic values with access to the original panic value

## Documentation

For comprehensive details on all functions and types, please refer to the full documentation available at:

[https://pkg.go.dev/github.com/mdobak/go-xerrors](https://pkg.go.dev/github.com/mdobak/go-xerrors)

## License

Licensed under the MIT License.
