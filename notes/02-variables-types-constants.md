# Chapter 02: Variables, Types and Constants

## What you will learn

- Declaring variables with `var` and `:=`
- Basic types: `int`, `float64`, `string`, `bool`
- Converting between types
- Constants and `iota`

## Key ideas

### Declaring variables

```go
var age int = 30
var name = "Sam"      // type is guessed
city := "Lagos"       // short form, only inside functions
var count int         // zero value is 0
```

Every type has a **zero value**: `0` for numbers, `""` for strings, `false` for bool,
`nil` for pointers, slices, maps and more.

### Types do not mix

Go never converts numbers for you. You must be explicit.

```go
var a int = 10
var b float64 = 2.5
sum := float64(a) + b // 12.5
```

### Constants and iota

```go
const Pi = 3.14159

type Color int

const (
	Red Color = iota // 0
	Green            // 1
	Blue             // 2
)
```

You can add a `String()` method to a type. Then `fmt.Println` will use it.

```go
func (c Color) String() string {
	switch c {
	case Red:
		return "Red"
	case Green:
		return "Green"
	case Blue:
		return "Blue"
	}
	return "Unknown"
}
```

---

## Exercises

### Exercise 1: Temperature

Write two functions.

- `CToF(c float64) float64` uses `c*9/5 + 32`
- `FToC(f float64) float64` is the reverse

**Starter** `ch02/ex01/temp.go`

```go
package temp

func CToF(c float64) float64 {
	return 0 // TODO
}

func FToC(f float64) float64 {
	return 0 // TODO
}
```

**Test** `ch02/ex01/temp_test.go`

```go
package temp

import (
	"math"
	"testing"
)

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) < 0.0001
}

func TestCToF(t *testing.T) {
	tests := []struct{ c, f float64 }{
		{0, 32},
		{100, 212},
		{-40, -40},
	}
	for _, tc := range tests {
		if got := CToF(tc.c); !almostEqual(got, tc.f) {
			t.Errorf("CToF(%v) = %v, want %v", tc.c, got, tc.f)
		}
	}
}

func TestFToC(t *testing.T) {
	if got := FToC(212); !almostEqual(got, 100) {
		t.Errorf("FToC(212) = %v, want 100", got)
	}
	if got := FToC(32); !almostEqual(got, 0) {
		t.Errorf("FToC(32) = %v, want 0", got)
	}
}
```

<details>
<summary>Solution</summary>

```go
package temp

func CToF(c float64) float64 {
	return c*9/5 + 32
}

func FToC(f float64) float64 {
	return (f - 32) * 5 / 9
}
```

</details>

### Exercise 2: Days of the week with iota

Create a `Weekday` type with constants `Sunday` (0) to `Saturday` (6).

- `String()` returns the day name, for example `"Monday"`. Anything out of range returns `"Unknown"`.
- `IsWeekend()` is true for Saturday and Sunday.

**Starter** `ch02/ex02/weekday.go`

```go
package weekday

type Weekday int

// TODO: add the constants Sunday ... Saturday using iota

func (d Weekday) String() string {
	return "" // TODO
}

func (d Weekday) IsWeekend() bool {
	return false // TODO
}
```

**Test** `ch02/ex02/weekday_test.go`

```go
package weekday

import (
	"fmt"
	"testing"
)

func TestValues(t *testing.T) {
	if Sunday != 0 || Monday != 1 || Saturday != 6 {
		t.Fatalf("wrong values: Sunday=%d Monday=%d Saturday=%d", Sunday, Monday, Saturday)
	}
}

func TestString(t *testing.T) {
	tests := map[Weekday]string{
		Sunday:    "Sunday",
		Monday:    "Monday",
		Wednesday: "Wednesday",
		Saturday:  "Saturday",
		Weekday(9): "Unknown",
	}
	for day, want := range tests {
		if got := day.String(); got != want {
			t.Errorf("%d.String() = %q, want %q", int(day), got, want)
		}
	}
}

func TestPrintUsesString(t *testing.T) {
	if got := fmt.Sprint(Friday); got != "Friday" {
		t.Errorf("Sprint(Friday) = %q, want Friday", got)
	}
}

func TestIsWeekend(t *testing.T) {
	if !Saturday.IsWeekend() || !Sunday.IsWeekend() {
		t.Error("Saturday and Sunday should be weekend days")
	}
	if Monday.IsWeekend() || Friday.IsWeekend() {
		t.Error("Monday and Friday should not be weekend days")
	}
}
```

<details>
<summary>Solution</summary>

```go
package weekday

type Weekday int

const (
	Sunday Weekday = iota
	Monday
	Tuesday
	Wednesday
	Thursday
	Friday
	Saturday
)

var names = [...]string{
	"Sunday", "Monday", "Tuesday", "Wednesday",
	"Thursday", "Friday", "Saturday",
}

func (d Weekday) String() string {
	if d < Sunday || d > Saturday {
		return "Unknown"
	}
	return names[d]
}

func (d Weekday) IsWeekend() bool {
	return d == Saturday || d == Sunday
}
```

</details>

### Exercise 3: Type conversion

Write two functions.

- `Percent(part, total int) float64` returns `part / total * 100`. If `total` is `0`, return `0`.
- `ToInt8(n int) (int8, bool)` returns the value and `true` if it fits in an `int8` (-128 to 127). If it does not fit, return `0, false`.

> [!TIP]
> `part / total` with two ints throws away the decimals. Convert to `float64` first.

**Starter** `ch02/ex03/convert.go`

```go
package convert

func Percent(part, total int) float64 {
	return 0 // TODO
}

func ToInt8(n int) (int8, bool) {
	return 0, false // TODO
}
```

**Test** `ch02/ex03/convert_test.go`

```go
package convert

import (
	"math"
	"testing"
)

func TestPercent(t *testing.T) {
	if got := Percent(1, 4); math.Abs(got-25) > 0.0001 {
		t.Errorf("Percent(1, 4) = %v, want 25", got)
	}
	if got := Percent(1, 3); math.Abs(got-33.3333) > 0.001 {
		t.Errorf("Percent(1, 3) = %v, want about 33.33", got)
	}
	if got := Percent(5, 0); got != 0 {
		t.Errorf("Percent(5, 0) = %v, want 0", got)
	}
}

func TestToInt8(t *testing.T) {
	tests := []struct {
		in   int
		want int8
		ok   bool
	}{
		{0, 0, true},
		{127, 127, true},
		{-128, -128, true},
		{128, 0, false},
		{-129, 0, false},
	}
	for _, tc := range tests {
		got, ok := ToInt8(tc.in)
		if got != tc.want || ok != tc.ok {
			t.Errorf("ToInt8(%d) = (%d, %v), want (%d, %v)", tc.in, got, ok, tc.want, tc.ok)
		}
	}
}
```

<details>
<summary>Solution</summary>

```go
package convert

import "math"

func Percent(part, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(part) / float64(total) * 100
}

func ToInt8(n int) (int8, bool) {
	if n < math.MinInt8 || n > math.MaxInt8 {
		return 0, false
	}
	return int8(n), true
}
```

</details>

## Checkpoint

- [ ] I know the zero value of each basic type
- [ ] I can convert between `int` and `float64`
- [ ] I can build an enum with `iota` and a `String()` method
