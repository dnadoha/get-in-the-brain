# Chapter 10: Error Handling

## What you will learn

- Errors are plain values
- Sentinel errors and `errors.Is`
- Custom error types and `errors.As`
- Wrapping errors with `%w`
- Combining errors and recovering from panics

## Key ideas

### Errors are values

```go
f, err := os.Open("data.txt")
if err != nil {
	return fmt.Errorf("open data: %w", err)
}
defer f.Close()
```

Go has no `try` and `catch`. You check `err` right after the call. It looks repeated,
but it makes the path of errors clear.

### Sentinel errors

A package level error that callers can compare with.

```go
var ErrNotFound = errors.New("not found")

if errors.Is(err, ErrNotFound) {
	// handle missing item
}
```

### Custom error types

```go
type ValidationError struct {
	Field string
}

func (e *ValidationError) Error() string {
	return e.Field + " is invalid"
}

var ve *ValidationError
if errors.As(err, &ve) {
	fmt.Println(ve.Field)
}
```

### Wrapping

`%w` wraps an error and adds context. `errors.Is` and `errors.As` still find the inner error.

```go
return fmt.Errorf("load user %d: %w", id, ErrNotFound)
```

### panic and recover

A panic stops the normal flow. Use it only for bugs, never for normal errors.
`recover` inside a deferred function catches a panic. Web servers use this so one bad request
does not crash the whole program.

---

## Exercises

### Exercise 1: Sentinel errors

Write `FindUser(users map[string]string, id string) (string, error)`. It returns the name for an id.

- If `id` is empty, return `ErrEmptyID`.
- If the id is not in the map, return `ErrNotFound`.

**Starter** `ch10/ex01/find.go`

```go
package find

import "errors"

var (
	ErrEmptyID  = errors.New("empty id")
	ErrNotFound = errors.New("not found")
)

func FindUser(users map[string]string, id string) (string, error) {
	return "", nil // TODO
}
```

**Test** `ch10/ex01/find_test.go`

```go
package find

import (
	"errors"
	"testing"
)

func TestFindUser(t *testing.T) {
	users := map[string]string{"1": "Ana", "2": "Bo"}

	name, err := FindUser(users, "2")
	if err != nil || name != "Bo" {
		t.Errorf("FindUser(2) = (%q, %v), want (Bo, nil)", name, err)
	}

	_, err = FindUser(users, "9")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}

	_, err = FindUser(users, "")
	if !errors.Is(err, ErrEmptyID) {
		t.Errorf("err = %v, want ErrEmptyID", err)
	}
}
```

<details>
<summary>Solution</summary>

```go
package find

import "errors"

var (
	ErrEmptyID  = errors.New("empty id")
	ErrNotFound = errors.New("not found")
)

func FindUser(users map[string]string, id string) (string, error) {
	if id == "" {
		return "", ErrEmptyID
	}
	name, ok := users[id]
	if !ok {
		return "", ErrNotFound
	}
	return name, nil
}
```

</details>

### Exercise 2: Custom error type

Create a `ValidationError` with fields `Field` and `Message`.

- `Error()` returns `"<Field>: <Message>"`, for example `"name: is required"`.
- `ValidateUser(name string, age int) error` returns `*ValidationError` for the **first** problem it finds:
  - empty name gives Field `"name"`, Message `"is required"`
  - age below 0 or above 150 gives Field `"age"`, Message `"must be between 0 and 150"`
- If all is fine, return `nil`.

**Starter** `ch10/ex02/validate.go`

```go
package validate

type ValidationError struct {
	// TODO: Field and Message
}

// TODO: func (e *ValidationError) Error() string

func ValidateUser(name string, age int) error {
	return nil // TODO
}
```

**Test** `ch10/ex02/validate_test.go`

```go
package validate

import (
	"errors"
	"testing"
)

func TestValidateOK(t *testing.T) {
	if err := ValidateUser("Ana", 30); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateName(t *testing.T) {
	err := ValidateUser("", 30)
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("err = %v, want *ValidationError", err)
	}
	if ve.Field != "name" || ve.Message != "is required" {
		t.Errorf("got field=%q message=%q", ve.Field, ve.Message)
	}
	if err.Error() != "name: is required" {
		t.Errorf("Error() = %q, want %q", err.Error(), "name: is required")
	}
}

func TestValidateAge(t *testing.T) {
	for _, age := range []int{-1, 151} {
		err := ValidateUser("Ana", age)
		var ve *ValidationError
		if !errors.As(err, &ve) {
			t.Fatalf("age %d: err = %v, want *ValidationError", age, err)
		}
		if ve.Field != "age" {
			t.Errorf("age %d: Field = %q, want age", age, ve.Field)
		}
		if err.Error() != "age: must be between 0 and 150" {
			t.Errorf("Error() = %q", err.Error())
		}
	}
}

func TestValidateFirstProblemWins(t *testing.T) {
	err := ValidateUser("", -5)
	var ve *ValidationError
	if !errors.As(err, &ve) || ve.Field != "name" {
		t.Errorf("expected the name error first, got %v", err)
	}
}
```

<details>
<summary>Solution</summary>

```go
package validate

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

func ValidateUser(name string, age int) error {
	if name == "" {
		return &ValidationError{Field: "name", Message: "is required"}
	}
	if age < 0 || age > 150 {
		return &ValidationError{Field: "age", Message: "must be between 0 and 150"}
	}
	return nil
}
```

</details>

### Exercise 3: Wrapping errors

Write `LoadUser(find func(id string) (string, error), id string) (string, error)`.

- Call `find(id)`.
- If it fails, return a **wrapped** error with the text `load user "<id>": <original error>`. Use `%w`.
- Callers must still be able to find the original error with `errors.Is`.

**Starter** `ch10/ex03/load.go`

```go
package load

func LoadUser(find func(id string) (string, error), id string) (string, error) {
	return "", nil // TODO
}
```

**Test** `ch10/ex03/load_test.go`

```go
package load

import (
	"errors"
	"testing"
)

var errNotFound = errors.New("not found")

func TestLoadUserOK(t *testing.T) {
	find := func(id string) (string, error) { return "Ana", nil }
	name, err := LoadUser(find, "1")
	if err != nil || name != "Ana" {
		t.Errorf("LoadUser = (%q, %v), want (Ana, nil)", name, err)
	}
}

func TestLoadUserWraps(t *testing.T) {
	find := func(id string) (string, error) { return "", errNotFound }
	_, err := LoadUser(find, "42")
	if err == nil {
		t.Fatal("expected an error")
	}
	if !errors.Is(err, errNotFound) {
		t.Error("errors.Is should find the original error: did you use %w?")
	}
	want := `load user "42": not found`
	if err.Error() != want {
		t.Errorf("message = %q, want %q", err.Error(), want)
	}
}
```

<details>
<summary>Solution</summary>

```go
package load

import "fmt"

func LoadUser(find func(id string) (string, error), id string) (string, error) {
	name, err := find(id)
	if err != nil {
		return "", fmt.Errorf("load user %q: %w", id, err)
	}
	return name, nil
}
```

</details>

### Exercise 4: Many errors at once

Write `CheckUser(name, email string, age int) error` that finds **all** problems, not just the first.

- Empty name adds `ErrNameRequired`
- Email without an `@` adds `ErrEmailInvalid`
- Age below 18 adds `ErrTooYoung`
- Return `nil` if there are no problems, otherwise `errors.Join(...)` of all of them

**Starter** `ch10/ex04/check.go`

```go
package check

import "errors"

var (
	ErrNameRequired = errors.New("name is required")
	ErrEmailInvalid = errors.New("email is invalid")
	ErrTooYoung     = errors.New("must be at least 18")
)

func CheckUser(name, email string, age int) error {
	return nil // TODO
}
```

**Test** `ch10/ex04/check_test.go`

```go
package check

import (
	"errors"
	"testing"
)

func TestCheckUserOK(t *testing.T) {
	if err := CheckUser("Ana", "ana@example.com", 30); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCheckUserOneProblem(t *testing.T) {
	err := CheckUser("Ana", "not-an-email", 30)
	if !errors.Is(err, ErrEmailInvalid) {
		t.Errorf("err = %v, want ErrEmailInvalid", err)
	}
	if errors.Is(err, ErrNameRequired) || errors.Is(err, ErrTooYoung) {
		t.Errorf("err = %v has problems that should not be there", err)
	}
}

func TestCheckUserAllProblems(t *testing.T) {
	err := CheckUser("", "nope", 10)
	for _, target := range []error{ErrNameRequired, ErrEmailInvalid, ErrTooYoung} {
		if !errors.Is(err, target) {
			t.Errorf("err = %v, should contain %v", err, target)
		}
	}
}
```

<details>
<summary>Solution</summary>

```go
package check

import (
	"errors"
	"strings"
)

var (
	ErrNameRequired = errors.New("name is required")
	ErrEmailInvalid = errors.New("email is invalid")
	ErrTooYoung     = errors.New("must be at least 18")
)

func CheckUser(name, email string, age int) error {
	var errs []error
	if name == "" {
		errs = append(errs, ErrNameRequired)
	}
	if !strings.Contains(email, "@") {
		errs = append(errs, ErrEmailInvalid)
	}
	if age < 18 {
		errs = append(errs, ErrTooYoung)
	}
	return errors.Join(errs...)
}
```

`errors.Join` returns `nil` when it gets no errors, so this works for the success case too.

</details>

### Exercise 5: Turn a panic into an error

Write `Safe(fn func()) (err error)`. It runs `fn`.

- If `fn` finishes normally, return `nil`.
- If `fn` panics, return an error with the text `recovered: <panic value>`.

> [!TIP]
> Use a named result `err`, and a `defer` with `recover()`.

**Starter** `ch10/ex05/safe.go`

```go
package safe

func Safe(fn func()) (err error) {
	return nil // TODO
}
```

**Test** `ch10/ex05/safe_test.go`

```go
package safe

import "testing"

func TestSafeNoPanic(t *testing.T) {
	called := false
	if err := Safe(func() { called = true }); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !called {
		t.Error("fn was not called")
	}
}

func TestSafePanicString(t *testing.T) {
	err := Safe(func() { panic("boom") })
	if err == nil {
		t.Fatal("expected an error")
	}
	if err.Error() != "recovered: boom" {
		t.Errorf("message = %q, want %q", err.Error(), "recovered: boom")
	}
}

func TestSafeRuntimePanic(t *testing.T) {
	err := Safe(func() {
		var m map[string]int
		m["a"] = 1 // writing to a nil map panics
	})
	if err == nil {
		t.Error("expected an error from a runtime panic")
	}
}
```

<details>
<summary>Solution</summary>

```go
package safe

import "fmt"

func Safe(fn func()) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("recovered: %v", r)
		}
	}()
	fn()
	return nil
}
```

</details>

## Checkpoint

- [ ] I check `err` right after every call that can fail
- [ ] I use `%w` to add context and still keep the original error
- [ ] I know when to use `errors.Is` and when to use `errors.As`
