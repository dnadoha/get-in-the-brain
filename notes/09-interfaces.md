# Chapter 09: Interfaces

## What you will learn

- Interfaces describe behavior, not data
- Types satisfy interfaces without saying so
- The empty interface `any` and type switches
- Why small interfaces make code easy to test

## Key ideas

### An interface is a set of methods

```go
type Shape interface {
	Area() float64
}

type Square struct{ Side float64 }

func (s Square) Area() float64 { return s.Side * s.Side }
```

`Square` has an `Area()` method, so it **is** a `Shape`. There is no `implements` keyword.

```go
func PrintArea(s Shape) {
	fmt.Println(s.Area())
}

PrintArea(Square{3})
```

### Useful interfaces from the standard library

| Interface | Method | Used by |
|-----------|--------|---------|
| `fmt.Stringer` | `String() string` | `fmt.Println` |
| `error` | `Error() string` | all error handling |
| `io.Reader` | `Read([]byte) (int, error)` | files, network, strings |
| `io.Writer` | `Write([]byte) (int, error)` | files, network, buffers |

### any and type switches

`any` accepts every value. A type switch finds out what it really is.

```go
func describe(v any) string {
	switch x := v.(type) {
	case int:
		return "int"
	case string:
		return "string of length " + strconv.Itoa(len(x))
	default:
		return "something else"
	}
}
```

### Accept interfaces, return structs

Make your functions depend on a small interface. In tests you can pass a simple fake.
You will use this idea again in the database chapter.

---

## Exercises

### Exercise 1: Shapes

Create two types that satisfy `Shape`:

- `Circle` with field `Radius float64`
- `Rect` with fields `Width, Height float64`

Both need `Area()` and `Perimeter()`. Then write `TotalArea(shapes []Shape) float64`.

Use `math.Pi` for the circle.

**Starter** `ch09/ex01/shape.go`

```go
package shape

type Shape interface {
	Area() float64
	Perimeter() float64
}

// TODO: type Circle and type Rect, with both methods

func TotalArea(shapes []Shape) float64 {
	return 0 // TODO
}
```

**Test** `ch09/ex01/shape_test.go`

```go
package shape

import (
	"math"
	"testing"
)

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) < 0.0001
}

func TestRect(t *testing.T) {
	var s Shape = Rect{Width: 3, Height: 4}
	if got := s.Area(); !almostEqual(got, 12) {
		t.Errorf("Area = %v, want 12", got)
	}
	if got := s.Perimeter(); !almostEqual(got, 14) {
		t.Errorf("Perimeter = %v, want 14", got)
	}
}

func TestCircle(t *testing.T) {
	var s Shape = Circle{Radius: 2}
	if got := s.Area(); !almostEqual(got, math.Pi*4) {
		t.Errorf("Area = %v, want %v", got, math.Pi*4)
	}
	if got := s.Perimeter(); !almostEqual(got, 2*math.Pi*2) {
		t.Errorf("Perimeter = %v, want %v", got, 2*math.Pi*2)
	}
}

func TestTotalArea(t *testing.T) {
	shapes := []Shape{
		Rect{Width: 2, Height: 5},
		Circle{Radius: 1},
		Rect{Width: 1, Height: 1},
	}
	want := 10 + math.Pi + 1
	if got := TotalArea(shapes); !almostEqual(got, want) {
		t.Errorf("TotalArea = %v, want %v", got, want)
	}
	if got := TotalArea(nil); got != 0 {
		t.Errorf("TotalArea(nil) = %v, want 0", got)
	}
}
```

<details>
<summary>Solution</summary>

```go
package shape

import "math"

type Shape interface {
	Area() float64
	Perimeter() float64
}

type Circle struct{ Radius float64 }

func (c Circle) Area() float64      { return math.Pi * c.Radius * c.Radius }
func (c Circle) Perimeter() float64 { return 2 * math.Pi * c.Radius }

type Rect struct{ Width, Height float64 }

func (r Rect) Area() float64      { return r.Width * r.Height }
func (r Rect) Perimeter() float64 { return 2 * (r.Width + r.Height) }

func TotalArea(shapes []Shape) float64 {
	total := 0.0
	for _, s := range shapes {
		total += s.Area()
	}
	return total
}
```

</details>

### Exercise 2: Stringer

`Money` stores cents. Add a `String()` method.

- `Money(1250)` prints as `"$12.50"`
- `Money(5)` prints as `"$0.05"`
- `Money(0)` prints as `"$0.00"`
- `Money(-305)` prints as `"-$3.05"`

> [!TIP]
> `fmt.Sprintf("%d.%02d", dollars, cents)` pads the cents with a zero.

**Starter** `ch09/ex02/money.go`

```go
package money

type Money int // cents

func (m Money) String() string {
	return "" // TODO
}
```

**Test** `ch09/ex02/money_test.go`

```go
package money

import (
	"fmt"
	"testing"
)

func TestString(t *testing.T) {
	tests := []struct {
		in   Money
		want string
	}{
		{1250, "$12.50"},
		{5, "$0.05"},
		{0, "$0.00"},
		{100, "$1.00"},
		{-305, "-$3.05"},
		{123456, "$1234.56"},
	}
	for _, tc := range tests {
		if got := tc.in.String(); got != tc.want {
			t.Errorf("Money(%d).String() = %q, want %q", int(tc.in), got, tc.want)
		}
	}
}

func TestPrintUsesString(t *testing.T) {
	if got := fmt.Sprint(Money(999)); got != "$9.99" {
		t.Errorf("Sprint = %q, want $9.99", got)
	}
}
```

<details>
<summary>Solution</summary>

```go
package money

import "fmt"

type Money int // cents

func (m Money) String() string {
	sign := ""
	if m < 0 {
		sign = "-"
		m = -m
	}
	return fmt.Sprintf("%s$%d.%02d", sign, int(m)/100, int(m)%100)
}
```

</details>

### Exercise 3: Type switch

Write `Describe(v any) string`.

| Input | Result |
|-------|--------|
| `5` (int) | `"int: 5"` |
| `"hi"` (string) | `"string: \"hi\""` |
| `true` (bool) | `"bool: true"` |
| `[]int{1, 2, 3}` | `"slice of 3 ints"` |
| `nil` | `"nil"` |
| anything else | `"unknown"` |

**Starter** `ch09/ex03/describe.go`

```go
package describe

func Describe(v any) string {
	return "" // TODO
}
```

**Test** `ch09/ex03/describe_test.go`

```go
package describe

import "testing"

func TestDescribe(t *testing.T) {
	tests := []struct {
		in   any
		want string
	}{
		{5, "int: 5"},
		{-12, "int: -12"},
		{"hi", `string: "hi"`},
		{true, "bool: true"},
		{false, "bool: false"},
		{[]int{1, 2, 3}, "slice of 3 ints"},
		{[]int{}, "slice of 0 ints"},
		{nil, "nil"},
		{3.14, "unknown"},
		{struct{}{}, "unknown"},
	}
	for _, tc := range tests {
		if got := Describe(tc.in); got != tc.want {
			t.Errorf("Describe(%v) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
```

<details>
<summary>Solution</summary>

```go
package describe

import "fmt"

func Describe(v any) string {
	switch x := v.(type) {
	case nil:
		return "nil"
	case int:
		return fmt.Sprintf("int: %d", x)
	case string:
		return fmt.Sprintf("string: %q", x)
	case bool:
		return fmt.Sprintf("bool: %t", x)
	case []int:
		return fmt.Sprintf("slice of %d ints", len(x))
	default:
		return "unknown"
	}
}
```

</details>

### Exercise 4: Depend on an interface

You have a `Notifier` interface. Write `WelcomeUser(n Notifier, email string) error`.

- If `email` is empty, return an error and do **not** send anything.
- Otherwise call `n.Send(email, "Welcome!")`.
- If `Send` fails, return that error.

The test uses a fake notifier. This is how you test code without sending real emails.

**Starter** `ch09/ex04/notify.go`

```go
package notify

type Notifier interface {
	Send(to, message string) error
}

func WelcomeUser(n Notifier, email string) error {
	return nil // TODO
}
```

**Test** `ch09/ex04/notify_test.go`

```go
package notify

import (
	"errors"
	"testing"
)

type fakeNotifier struct {
	sentTo  []string
	sentMsg []string
	err     error
}

func (f *fakeNotifier) Send(to, message string) error {
	if f.err != nil {
		return f.err
	}
	f.sentTo = append(f.sentTo, to)
	f.sentMsg = append(f.sentMsg, message)
	return nil
}

func TestWelcomeUser(t *testing.T) {
	f := &fakeNotifier{}
	if err := WelcomeUser(f, "ana@example.com"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(f.sentTo) != 1 || f.sentTo[0] != "ana@example.com" || f.sentMsg[0] != "Welcome!" {
		t.Errorf("wrong message sent: to=%v msg=%v", f.sentTo, f.sentMsg)
	}
}

func TestWelcomeUserEmptyEmail(t *testing.T) {
	f := &fakeNotifier{}
	if err := WelcomeUser(f, ""); err == nil {
		t.Error("expected an error for an empty email")
	}
	if len(f.sentTo) != 0 {
		t.Error("nothing should be sent when the email is empty")
	}
}

func TestWelcomeUserSendFails(t *testing.T) {
	boom := errors.New("smtp down")
	f := &fakeNotifier{err: boom}
	if err := WelcomeUser(f, "ana@example.com"); err != boom {
		t.Errorf("err = %v, want %v", err, boom)
	}
}
```

<details>
<summary>Solution</summary>

```go
package notify

import "errors"

type Notifier interface {
	Send(to, message string) error
}

func WelcomeUser(n Notifier, email string) error {
	if email == "" {
		return errors.New("email is required")
	}
	return n.Send(email, "Welcome!")
}
```

</details>

## Checkpoint

- [ ] I know a type satisfies an interface by having the methods
- [ ] I can use a type switch on `any`
- [ ] I can write a fake to test code that depends on an interface
