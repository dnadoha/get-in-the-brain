# Chapter 11: Packages and Testing

## What you will learn

- How packages and modules are organized
- Exported and unexported names
- Table driven tests and subtests
- Reading a failing test
- Benchmarks and coverage

## Key ideas

### Packages and modules

- A **package** is a folder of `.go` files that share one `package` name.
- A **module** is a group of packages with one `go.mod` file at the root.
- Import another package with its full path: `import "golearn/ch11/ex04/calc"`.

```text
golearn/
  go.mod              module golearn
  ch11/ex04/calc/
    calc.go           package calc
    calc_test.go
```

A name is **exported** if it starts with a capital letter. Everything else is private to the package.
A folder named `internal` can only be imported by code inside its parent folder. Use it to hide
code from other modules.

### Table driven tests

The most common Go test style. List the cases in a slice, then loop.

```go
func TestAbs(t *testing.T) {
	tests := []struct {
		name string
		in   int
		want int
	}{
		{"positive", 5, 5},
		{"negative", -5, 5},
		{"zero", 0, 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := Abs(tc.in); got != tc.want {
				t.Errorf("Abs(%d) = %d, want %d", tc.in, got, tc.want)
			}
		})
	}
}
```

`t.Run` creates a **subtest**. You can run just one with:

```bash
go test -run 'TestAbs/negative' -v ./...
```

### Tips for good tests

- Name the case that failed, and print both `got` and `want`.
- Use `t.Helper()` inside helper functions so the error points to the caller.
- Use `t.TempDir()` for files. Go deletes the folder for you.
- Put tests in `package calc_test` (an external test package) to test only the public API.

### Benchmarks

```go
func BenchmarkAbs(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Abs(-5)
	}
}
```

```bash
go test -bench=. -benchmem
```

### Coverage

```bash
go test -cover ./...
go test -coverprofile=c.out ./... && go tool cover -html=c.out
```

---

## Exercises

### Exercise 1: Pass a table test

Implement two functions so the table tests pass.

- `Clamp(v, lo, hi int) int` keeps `v` between `lo` and `hi`.
- `Slugify(s string) string` makes a URL friendly text:
  - lower case letters and digits are kept
  - every group of other characters becomes a single `-`
  - no `-` at the start or the end

**Starter** `ch11/ex01/util.go`

```go
package util

func Clamp(v, lo, hi int) int {
	return 0 // TODO
}

func Slugify(s string) string {
	return "" // TODO
}
```

**Test** `ch11/ex01/util_test.go`

```go
package util

import "testing"

func TestClamp(t *testing.T) {
	tests := []struct {
		name         string
		v, lo, hi    int
		want         int
	}{
		{"inside", 5, 0, 10, 5},
		{"below", -3, 0, 10, 0},
		{"above", 42, 0, 10, 10},
		{"at low edge", 0, 0, 10, 0},
		{"at high edge", 10, 0, 10, 10},
		{"negative range", -7, -10, -5, -7},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := Clamp(tc.v, tc.lo, tc.hi); got != tc.want {
				t.Errorf("Clamp(%d, %d, %d) = %d, want %d", tc.v, tc.lo, tc.hi, got, tc.want)
			}
		})
	}
}

func TestSlugify(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", ""},
		{"simple", "Hello World", "hello-world"},
		{"punctuation", "Hello, World!", "hello-world"},
		{"extra spaces", "  Go  is   FUN ", "go-is-fun"},
		{"repeated dashes", "a--b", "a-b"},
		{"only symbols", "***", ""},
		{"digits", "Go 1.22", "go-1-22"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := Slugify(tc.in); got != tc.want {
				t.Errorf("Slugify(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
```

<details>
<summary>Solution</summary>

```go
package util

import (
	"strings"
	"unicode"
)

func Clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func Slugify(s string) string {
	var b strings.Builder
	lastDash := true // avoids a dash at the start
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(unicode.ToLower(r))
			lastDash = false
		} else if !lastDash {
			b.WriteByte('-')
			lastDash = true
		}
	}
	return strings.TrimSuffix(b.String(), "-")
}
```

</details>

### Exercise 2: Find the bug

The code below has a bug. Run the test, read the failing subtest names, find the bug and fix it.

A year is a leap year if it is divisible by 4, **except** years divisible by 100, **unless** they are also divisible by 400.

**Starter** `ch11/ex02/leap.go`

```go
package leap

func IsLeapYear(year int) bool {
	return year%4 == 0 && year%100 != 0
}

func DaysInMonth(year, month int) int {
	switch month {
	case 1, 3, 5, 7, 8, 10, 12:
		return 31
	case 4, 6, 9, 11:
		return 30
	case 2:
		if IsLeapYear(year) {
			return 29
		}
		return 28
	}
	return 0
}
```

**Test** `ch11/ex02/leap_test.go`

```go
package leap

import (
	"fmt"
	"testing"
)

func TestIsLeapYear(t *testing.T) {
	tests := []struct {
		year int
		want bool
	}{
		{2023, false},
		{2024, true},
		{1900, false},
		{2000, true},
		{2100, false},
		{1600, true},
		{1, false},
	}
	for _, tc := range tests {
		t.Run(fmt.Sprint(tc.year), func(t *testing.T) {
			if got := IsLeapYear(tc.year); got != tc.want {
				t.Errorf("IsLeapYear(%d) = %v, want %v", tc.year, got, tc.want)
			}
		})
	}
}

func TestDaysInMonth(t *testing.T) {
	tests := []struct {
		year, month, want int
	}{
		{2023, 1, 31},
		{2023, 2, 28},
		{2024, 2, 29},
		{2000, 2, 29},
		{1900, 2, 28},
		{2023, 4, 30},
		{2023, 12, 31},
		{2023, 0, 0},
		{2023, 13, 0},
	}
	for _, tc := range tests {
		name := fmt.Sprintf("%d-%d", tc.year, tc.month)
		t.Run(name, func(t *testing.T) {
			if got := DaysInMonth(tc.year, tc.month); got != tc.want {
				t.Errorf("DaysInMonth(%d, %d) = %d, want %d", tc.year, tc.month, got, tc.want)
			}
		})
	}
}
```

<details>
<summary>Solution</summary>

```go
func IsLeapYear(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}
```

The first version forgot the rule for years divisible by 400, so `2000` and `1600` failed.
`DaysInMonth` was already correct. It failed for `2000` only because it uses `IsLeapYear`.

</details>

### Exercise 3: Write and run a benchmark

- Write `JoinWords(words []string) string`. It joins the words with one space. Use `strings.Builder`.
- Then run the benchmark and look at the numbers:

```bash
go test -bench=. -benchmem ./ch11/ex03/
```

The benchmark compares your function with a slow version that uses `+=`. You should see fewer
allocations (`allocs/op`) for the builder.

**Starter** `ch11/ex03/join.go`

```go
package join

func JoinWords(words []string) string {
	return "" // TODO: use strings.Builder
}

// joinSlow is here for the benchmark. Do not change it.
func joinSlow(words []string) string {
	out := ""
	for i, w := range words {
		if i > 0 {
			out += " "
		}
		out += w
	}
	return out
}
```

**Test** `ch11/ex03/join_test.go`

```go
package join

import (
	"strings"
	"testing"
)

func TestJoinWords(t *testing.T) {
	tests := []struct {
		name  string
		words []string
		want  string
	}{
		{"empty", nil, ""},
		{"one", []string{"go"}, "go"},
		{"many", []string{"go", "is", "fun"}, "go is fun"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := JoinWords(tc.words); got != tc.want {
				t.Errorf("JoinWords(%v) = %q, want %q", tc.words, got, tc.want)
			}
		})
	}
}

func makeWords(n int) []string {
	words := make([]string, n)
	for i := range words {
		words[i] = "word"
	}
	return words
}

func TestSameAsStringsJoin(t *testing.T) {
	words := makeWords(50)
	if got, want := JoinWords(words), strings.Join(words, " "); got != want {
		t.Errorf("result differs from strings.Join")
	}
}

func BenchmarkJoinWords(b *testing.B) {
	words := makeWords(1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		JoinWords(words)
	}
}

func BenchmarkJoinSlow(b *testing.B) {
	words := makeWords(1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		joinSlow(words)
	}
}
```

<details>
<summary>Solution</summary>

```go
func JoinWords(words []string) string {
	var b strings.Builder
	for i, w := range words {
		if i > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(w)
	}
	return b.String()
}
```

Add `import "strings"` at the top of the file.

</details>

### Exercise 4: Your own package, tested from outside

This time you work with **two** packages and an external test package.

Implement `Mean` in the package `calc`. Use the unexported helper `sum` that is already there.

- `Mean(nums []float64) (float64, error)` returns the average.
- If `nums` is empty, return `0` and `ErrNoData`.

The test is in `package calc_test`, so it can only use exported names.

**Starter** `ch11/ex04/calc/calc.go`

```go
package calc

import "errors"

var ErrNoData = errors.New("no data")

// sum is unexported: code outside this package cannot call it.
func sum(nums []float64) float64 {
	total := 0.0
	for _, n := range nums {
		total += n
	}
	return total
}

func Mean(nums []float64) (float64, error) {
	return 0, nil // TODO
}
```

**Test** `ch11/ex04/calc/calc_test.go`

```go
package calc_test

import (
	"errors"
	"math"
	"testing"

	"golearn/ch11/ex04/calc"
)

func TestMean(t *testing.T) {
	got, err := calc.Mean([]float64{2, 4, 6, 8})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if math.Abs(got-5) > 0.0001 {
		t.Errorf("Mean = %v, want 5", got)
	}
}

func TestMeanSingleValue(t *testing.T) {
	got, err := calc.Mean([]float64{7.5})
	if err != nil || got != 7.5 {
		t.Errorf("Mean = (%v, %v), want (7.5, nil)", got, err)
	}
}

func TestMeanEmpty(t *testing.T) {
	got, err := calc.Mean(nil)
	if !errors.Is(err, calc.ErrNoData) {
		t.Errorf("err = %v, want ErrNoData", err)
	}
	if got != 0 {
		t.Errorf("Mean(nil) = %v, want 0", got)
	}
}
```

<details>
<summary>Solution</summary>

```go
func Mean(nums []float64) (float64, error) {
	if len(nums) == 0 {
		return 0, ErrNoData
	}
	return sum(nums) / float64(len(nums)), nil
}
```

</details>

> [!NOTE]
> The import path `golearn/ch11/ex04/calc` works because your module is called `golearn`.
> If you used another module name, change the path in the test.

## Checkpoint

- [ ] I can write a table driven test with `t.Run`
- [ ] I can run a single subtest with `-run`
- [ ] I know the difference between an exported and an unexported name
- [ ] I can run a benchmark and read `allocs/op`
