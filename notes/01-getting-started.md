# Chapter 01: Getting Started

## What you will learn

- What a Go module is and why `go.mod` exists
- How to run, build, format and test code
- How to read a Go test
- How `package` and `import` work

## Key ideas

### Your first program

Create `hello/main.go`:

```go
package main

import "fmt"

func main() {
	fmt.Println("Hello, Go!")
}
```

Run it:

```bash
go run ./hello
```

- `package main` tells Go this is a program you can run.
- `func main()` is where the program starts.
- `fmt` is a package from the standard library.

### The tools you get for free

| Command | Purpose |
|---------|---------|
| `go run` | Compile and run in one step |
| `go build` | Create a binary file |
| `go fmt` / `gofmt` | Format code the Go way |
| `go vet` | Find suspicious code |
| `go test` | Run tests |

### How a test looks

A test lives in a file that ends with `_test.go`. The function name starts with `Test`
and takes `*testing.T`.

```go
func TestAdd(t *testing.T) {
	got := Add(2, 3)
	want := 5
	if got != want {
		t.Errorf("Add(2, 3) = %d, want %d", got, want)
	}
}
```

Use `t.Errorf` to report a failure and keep going. Use `t.Fatalf` to stop the test at once.

---

## Exercises

### Exercise 1: Greet someone

Write `Greet(name string) string`.

- `Greet("Ana")` returns `"Hello, Ana!"`
- `Greet("")` returns `"Hello, World!"`

**Starter** `ch01/ex01/greet.go`

```go
package greet

// Greet returns a greeting for the given name.
func Greet(name string) string {
	return "" // TODO
}
```

**Test** `ch01/ex01/greet_test.go`

```go
package greet

import "testing"

func TestGreet(t *testing.T) {
	if got := Greet("Ana"); got != "Hello, Ana!" {
		t.Errorf("Greet(\"Ana\") = %q, want %q", got, "Hello, Ana!")
	}
}

func TestGreetEmpty(t *testing.T) {
	if got := Greet(""); got != "Hello, World!" {
		t.Errorf("Greet(\"\") = %q, want %q", got, "Hello, World!")
	}
}
```

<details>
<summary>Solution</summary>

```go
package greet

func Greet(name string) string {
	if name == "" {
		name = "World"
	}
	return "Hello, " + name + "!"
}
```

</details>

### Exercise 2: Small math helpers

Write two functions:

- `Add(a, b int) int`
- `Abs(n int) int` returns the number without its minus sign

**Starter** `ch01/ex02/mathx.go`

```go
package mathx

func Add(a, b int) int {
	return 0 // TODO
}

func Abs(n int) int {
	return 0 // TODO
}
```

**Test** `ch01/ex02/mathx_test.go`

```go
package mathx

import "testing"

func TestAdd(t *testing.T) {
	if got := Add(2, 3); got != 5 {
		t.Errorf("Add(2, 3) = %d, want 5", got)
	}
	if got := Add(-4, 1); got != -3 {
		t.Errorf("Add(-4, 1) = %d, want -3", got)
	}
}

func TestAbs(t *testing.T) {
	if got := Abs(-7); got != 7 {
		t.Errorf("Abs(-7) = %d, want 7", got)
	}
	if got := Abs(7); got != 7 {
		t.Errorf("Abs(7) = %d, want 7", got)
	}
	if got := Abs(0); got != 0 {
		t.Errorf("Abs(0) = %d, want 0", got)
	}
}
```

<details>
<summary>Solution</summary>

```go
package mathx

func Add(a, b int) int {
	return a + b
}

func Abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
```

</details>

### Exercise 3: Write your own test

This time the function is done. **You** write the test.

Create `ch01/ex03/double.go`:

```go
package double

func Double(n int) int {
	return n * 2
}
```

Then create `ch01/ex03/double_test.go` with a `TestDouble` that checks at least three
values: a positive number, zero, and a negative number. Run `go test -v ./ch01/ex03/`.

<details>
<summary>Possible answer</summary>

```go
package double

import "testing"

func TestDouble(t *testing.T) {
	if got := Double(4); got != 8 {
		t.Errorf("Double(4) = %d, want 8", got)
	}
	if got := Double(0); got != 0 {
		t.Errorf("Double(0) = %d, want 0", got)
	}
	if got := Double(-3); got != -6 {
		t.Errorf("Double(-3) = %d, want -6", got)
	}
}
```

</details>

## Checkpoint

- [ ] I can run a program with `go run`
- [ ] I know what `go.mod` is for
- [ ] I can write a simple test
