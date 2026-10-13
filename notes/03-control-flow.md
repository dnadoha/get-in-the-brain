# Chapter 03: Control Flow

## What you will learn

- `if` and `else`, with a short statement
- `switch` without fallthrough
- `for`, the only loop in Go
- `break` and `continue`

## Key ideas

### if

No parentheses, braces are required.

```go
if n%2 == 0 {
	fmt.Println("even")
} else {
	fmt.Println("odd")
}

// A short statement before the condition
if v, err := strconv.Atoi("42"); err == nil {
	fmt.Println(v)
}
```

### switch

Cases do not fall through. You do not need `break`.

```go
switch {
case score >= 90:
	return "A"
case score >= 80:
	return "B"
default:
	return "F"
}
```

### for

```go
for i := 0; i < 3; i++ { }   // classic
for n < 100 { n *= 2 }       // like while
for { break }                // forever
for i := range 3 { }         // 0, 1, 2 (Go 1.22+)
```

---

## Exercises

### Exercise 1: FizzBuzz

Write `FizzBuzz(n int) string`.

- Divisible by 3 and 5: `"FizzBuzz"`
- Divisible by 3: `"Fizz"`
- Divisible by 5: `"Buzz"`
- Otherwise the number as text, for example `"7"`

> [!TIP]
> `strconv.Itoa(7)` turns `7` into `"7"`.

**Starter** `ch03/ex01/fizz.go`

```go
package fizz

func FizzBuzz(n int) string {
	return "" // TODO
}
```

**Test** `ch03/ex01/fizz_test.go`

```go
package fizz

import "testing"

func TestFizzBuzz(t *testing.T) {
	tests := []struct {
		in   int
		want string
	}{
		{1, "1"},
		{2, "2"},
		{3, "Fizz"},
		{5, "Buzz"},
		{7, "7"},
		{9, "Fizz"},
		{10, "Buzz"},
		{15, "FizzBuzz"},
		{30, "FizzBuzz"},
	}
	for _, tc := range tests {
		if got := FizzBuzz(tc.in); got != tc.want {
			t.Errorf("FizzBuzz(%d) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
```

<details>
<summary>Solution</summary>

```go
package fizz

import "strconv"

func FizzBuzz(n int) string {
	switch {
	case n%15 == 0:
		return "FizzBuzz"
	case n%3 == 0:
		return "Fizz"
	case n%5 == 0:
		return "Buzz"
	}
	return strconv.Itoa(n)
}
```

</details>

### Exercise 2: Grades

Write `Grade(score int) string`.

| Score | Result |
|-------|--------|
| 90 to 100 | `"A"` |
| 80 to 89 | `"B"` |
| 70 to 79 | `"C"` |
| 60 to 69 | `"D"` |
| 0 to 59 | `"F"` |
| below 0 or above 100 | `"Invalid"` |

**Starter** `ch03/ex02/grade.go`

```go
package grade

func Grade(score int) string {
	return "" // TODO
}
```

**Test** `ch03/ex02/grade_test.go`

```go
package grade

import "testing"

func TestGrade(t *testing.T) {
	tests := []struct {
		score int
		want  string
	}{
		{100, "A"}, {90, "A"},
		{89, "B"}, {80, "B"},
		{79, "C"}, {70, "C"},
		{69, "D"}, {60, "D"},
		{59, "F"}, {0, "F"},
		{-1, "Invalid"}, {101, "Invalid"},
	}
	for _, tc := range tests {
		if got := Grade(tc.score); got != tc.want {
			t.Errorf("Grade(%d) = %q, want %q", tc.score, got, tc.want)
		}
	}
}
```

<details>
<summary>Solution</summary>

```go
package grade

func Grade(score int) string {
	switch {
	case score < 0 || score > 100:
		return "Invalid"
	case score >= 90:
		return "A"
	case score >= 80:
		return "B"
	case score >= 70:
		return "C"
	case score >= 60:
		return "D"
	}
	return "F"
}
```

</details>

### Exercise 3: Prime numbers

Write `IsPrime(n int) bool`. Numbers below 2 are not prime.

> [!TIP]
> You only need to check divisors up to the square root of `n`. Try `for i := 2; i*i <= n; i++`.

**Starter** `ch03/ex03/prime.go`

```go
package prime

func IsPrime(n int) bool {
	return false // TODO
}
```

**Test** `ch03/ex03/prime_test.go`

```go
package prime

import "testing"

func TestIsPrime(t *testing.T) {
	primes := []int{2, 3, 5, 7, 11, 13, 97, 7919}
	notPrimes := []int{-5, 0, 1, 4, 6, 9, 15, 100, 7917}

	for _, n := range primes {
		if !IsPrime(n) {
			t.Errorf("IsPrime(%d) = false, want true", n)
		}
	}
	for _, n := range notPrimes {
		if IsPrime(n) {
			t.Errorf("IsPrime(%d) = true, want false", n)
		}
	}
}
```

<details>
<summary>Solution</summary>

```go
package prime

func IsPrime(n int) bool {
	if n < 2 {
		return false
	}
	for i := 2; i*i <= n; i++ {
		if n%i == 0 {
			return false
		}
	}
	return true
}
```

</details>

### Exercise 4: Collatz steps

Start with a number `n`. If it is even, divide by 2. If it is odd, do `3*n + 1`.
Repeat until you reach `1`. Write `CollatzSteps(n int) int` that returns how many steps it took.

- `CollatzSteps(1)` is `0`
- `CollatzSteps(6)` is `8` (6, 3, 10, 5, 16, 8, 4, 2, 1)
- If `n` is `0` or negative, return `-1`

**Starter** `ch03/ex04/collatz.go`

```go
package collatz

func CollatzSteps(n int) int {
	return 0 // TODO
}
```

**Test** `ch03/ex04/collatz_test.go`

```go
package collatz

import "testing"

func TestCollatzSteps(t *testing.T) {
	tests := []struct{ in, want int }{
		{1, 0},
		{2, 1},
		{6, 8},
		{7, 16},
		{27, 111},
		{0, -1},
		{-3, -1},
	}
	for _, tc := range tests {
		if got := CollatzSteps(tc.in); got != tc.want {
			t.Errorf("CollatzSteps(%d) = %d, want %d", tc.in, got, tc.want)
		}
	}
}
```

<details>
<summary>Solution</summary>

```go
package collatz

func CollatzSteps(n int) int {
	if n <= 0 {
		return -1
	}
	steps := 0
	for n != 1 {
		if n%2 == 0 {
			n /= 2
		} else {
			n = 3*n + 1
		}
		steps++
	}
	return steps
}
```

</details>

## Checkpoint

- [ ] I can write all forms of `for`
- [ ] I know `switch` does not fall through
- [ ] I can use `switch` with no value for ranges
