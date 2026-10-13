# Chapter 04: Functions

## What you will learn

- Parameters and multiple return values
- Variadic functions
- Functions as values and closures
- `defer`

## Key ideas

### Multiple return values

Go functions often return a result and an error.

```go
func Divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, errors.New("division by zero")
	}
	return a / b, nil
}

result, err := Divide(10, 2)
if err != nil {
	// handle it
}
```

### Variadic functions

```go
func Sum(nums ...int) int {
	total := 0
	for _, n := range nums {
		total += n
	}
	return total
}

Sum(1, 2, 3)
```

Inside the function, `nums` is a slice of `int`. You will learn slices in the next chapter.

### Functions are values

```go
func apply(n int, f func(int) int) int { return f(n) }

double := func(x int) int { return x * 2 }
apply(5, double) // 10
```

### Closures

A function can remember variables from the place where it was created.

```go
func counter() func() int {
	n := 0
	return func() int {
		n++
		return n
	}
}
```

### defer

`defer` runs a call when the surrounding function ends. Deferred calls run in
**last in, first out** order. This is how Go closes files and unlocks mutexes.

```go
f, err := os.Open("data.txt")
if err != nil {
	return err
}
defer f.Close()
```

---

## Exercises

### Exercise 1: Divide and MinMax

- `Divide(a, b float64) (float64, error)` returns an error when `b` is `0`.
- `MinMax(a, b, c int) (min, max int)` returns the smallest and the largest value. Use named results.

**Starter** `ch04/ex01/funcs.go`

```go
package funcs

func Divide(a, b float64) (float64, error) {
	return 0, nil // TODO
}

func MinMax(a, b, c int) (min, max int) {
	return // TODO
}
```

**Test** `ch04/ex01/funcs_test.go`

```go
package funcs

import "testing"

func TestDivide(t *testing.T) {
	got, err := Divide(10, 4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 2.5 {
		t.Errorf("Divide(10, 4) = %v, want 2.5", got)
	}
}

func TestDivideByZero(t *testing.T) {
	_, err := Divide(1, 0)
	if err == nil {
		t.Error("expected an error when dividing by zero")
	}
}

func TestMinMax(t *testing.T) {
	tests := []struct{ a, b, c, min, max int }{
		{1, 2, 3, 1, 3},
		{3, 2, 1, 1, 3},
		{2, 9, -4, -4, 9},
		{5, 5, 5, 5, 5},
	}
	for _, tc := range tests {
		min, max := MinMax(tc.a, tc.b, tc.c)
		if min != tc.min || max != tc.max {
			t.Errorf("MinMax(%d, %d, %d) = (%d, %d), want (%d, %d)",
				tc.a, tc.b, tc.c, min, max, tc.min, tc.max)
		}
	}
}
```

<details>
<summary>Solution</summary>

```go
package funcs

import "errors"

func Divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, errors.New("division by zero")
	}
	return a / b, nil
}

func MinMax(a, b, c int) (min, max int) {
	min, max = a, a
	for _, v := range []int{b, c} {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	return
}
```

</details>

### Exercise 2: Variadic functions

- `Sum(nums ...int) int` adds all numbers. No numbers means `0`.
- `Max(first int, rest ...int) int` returns the biggest number. It needs at least one number, so the first one is a normal parameter.

**Starter** `ch04/ex02/variadic.go`

```go
package variadic

func Sum(nums ...int) int {
	return 0 // TODO
}

func Max(first int, rest ...int) int {
	return 0 // TODO
}
```

**Test** `ch04/ex02/variadic_test.go`

```go
package variadic

import "testing"

func TestSum(t *testing.T) {
	if got := Sum(); got != 0 {
		t.Errorf("Sum() = %d, want 0", got)
	}
	if got := Sum(1, 2, 3, 4); got != 10 {
		t.Errorf("Sum(1, 2, 3, 4) = %d, want 10", got)
	}
	nums := []int{5, 5, 5}
	if got := Sum(nums...); got != 15 {
		t.Errorf("Sum(nums...) = %d, want 15", got)
	}
}

func TestMax(t *testing.T) {
	if got := Max(7); got != 7 {
		t.Errorf("Max(7) = %d, want 7", got)
	}
	if got := Max(1, 9, 3); got != 9 {
		t.Errorf("Max(1, 9, 3) = %d, want 9", got)
	}
	if got := Max(-5, -2, -9); got != -2 {
		t.Errorf("Max(-5, -2, -9) = %d, want -2", got)
	}
}
```

<details>
<summary>Solution</summary>

```go
package variadic

func Sum(nums ...int) int {
	total := 0
	for _, n := range nums {
		total += n
	}
	return total
}

func Max(first int, rest ...int) int {
	best := first
	for _, n := range rest {
		if n > best {
			best = n
		}
	}
	return best
}
```

</details>

### Exercise 3: Closures

- `Counter() func() int` returns a function. Each call returns the next number, starting at `1`. Two counters must not share state.
- `Fibonacci() func() int` returns a function. Each call returns the next Fibonacci number: `0, 1, 1, 2, 3, 5, 8 ...`

**Starter** `ch04/ex03/closure.go`

```go
package closure

func Counter() func() int {
	return nil // TODO
}

func Fibonacci() func() int {
	return nil // TODO
}
```

**Test** `ch04/ex03/closure_test.go`

```go
package closure

import "testing"

func TestCounter(t *testing.T) {
	c := Counter()
	for want := 1; want <= 3; want++ {
		if got := c(); got != want {
			t.Errorf("call %d returned %d", want, got)
		}
	}
}

func TestCountersAreIndependent(t *testing.T) {
	a := Counter()
	b := Counter()
	a()
	a()
	if got := b(); got != 1 {
		t.Errorf("b() = %d, want 1 (counters must not share state)", got)
	}
}

func TestFibonacci(t *testing.T) {
	f := Fibonacci()
	want := []int{0, 1, 1, 2, 3, 5, 8, 13, 21, 34}
	for i, w := range want {
		if got := f(); got != w {
			t.Errorf("call %d = %d, want %d", i+1, got, w)
		}
	}
}
```

<details>
<summary>Solution</summary>

```go
package closure

func Counter() func() int {
	n := 0
	return func() int {
		n++
		return n
	}
}

func Fibonacci() func() int {
	a, b := 0, 1
	return func() int {
		result := a
		a, b = b, a+b
		return result
	}
}
```

</details>

### Exercise 4: Functions as values

- `Apply(nums []int, f func(int) int) []int` returns a new slice with `f` applied to every number. Do not change the input.
- `Compose(f, g func(int) int) func(int) int` returns a function that does `g(f(x))`.

**Starter** `ch04/ex04/higher.go`

```go
package higher

func Apply(nums []int, f func(int) int) []int {
	return nil // TODO
}

func Compose(f, g func(int) int) func(int) int {
	return nil // TODO
}
```

**Test** `ch04/ex04/higher_test.go`

```go
package higher

import (
	"reflect"
	"testing"
)

func TestApply(t *testing.T) {
	in := []int{1, 2, 3}
	got := Apply(in, func(x int) int { return x * x })
	want := []int{1, 4, 9}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Apply = %v, want %v", got, want)
	}
	if !reflect.DeepEqual(in, []int{1, 2, 3}) {
		t.Errorf("Apply changed its input: %v", in)
	}
}

func TestApplyEmpty(t *testing.T) {
	got := Apply([]int{}, func(x int) int { return x })
	if len(got) != 0 {
		t.Errorf("expected an empty result, got %v", got)
	}
}

func TestCompose(t *testing.T) {
	addOne := func(x int) int { return x + 1 }
	double := func(x int) int { return x * 2 }

	// first add one, then double
	h := Compose(addOne, double)
	if got := h(5); got != 12 {
		t.Errorf("Compose(addOne, double)(5) = %d, want 12", got)
	}
}
```

<details>
<summary>Solution</summary>

```go
package higher

func Apply(nums []int, f func(int) int) []int {
	out := make([]int, 0, len(nums))
	for _, n := range nums {
		out = append(out, f(n))
	}
	return out
}

func Compose(f, g func(int) int) func(int) int {
	return func(x int) int {
		return g(f(x))
	}
}
```

</details>

### Exercise 5: defer order

Write `Process(record func(string))`. It must call `record` with these texts, in this order of **calls in the code**:

1. `record("start")`
1. Use `defer` to register `record("first deferred")`
1. Use `defer` to register `record("second deferred")`
1. `record("end")`

The test checks the order in which the messages really arrive. Think about last in, first out.

**Starter** `ch04/ex05/deferred.go`

```go
package deferred

func Process(record func(string)) {
	// TODO
}
```

**Test** `ch04/ex05/deferred_test.go`

```go
package deferred

import (
	"reflect"
	"testing"
)

func TestProcess(t *testing.T) {
	var got []string
	Process(func(s string) { got = append(got, s) })

	want := []string{"start", "end", "second deferred", "first deferred"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("order = %v, want %v", got, want)
	}
}
```

<details>
<summary>Solution</summary>

```go
package deferred

func Process(record func(string)) {
	record("start")
	defer record("first deferred")
	defer record("second deferred")
	record("end")
}
```

</details>

## Checkpoint

- [ ] I can return more than one value
- [ ] I can write a closure that keeps its own state
- [ ] I can explain the order in which deferred calls run
