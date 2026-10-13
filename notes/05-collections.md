# Chapter 05: Arrays, Slices and Maps

## What you will learn

- The difference between arrays and slices
- `append`, `len`, `cap`, `copy`
- Why slices can surprise you
- Maps: create, read, delete, check if a key exists

## Key ideas

### Arrays and slices

An array has a fixed size. A slice is a flexible window over an array. You will use slices
almost all the time.

```go
var arr [3]int            // [0 0 0], size is part of the type
nums := []int{1, 2, 3}    // slice
nums = append(nums, 4)    // [1 2 3 4]
part := nums[1:3]         // [2 3], shares memory with nums
```

### The sharing trap

A sub slice shares memory with the original. Changing one changes the other.

```go
a := []int{1, 2, 3}
b := a[:2]
b[0] = 99
fmt.Println(a) // [99 2 3]
```

To get an independent copy, use `copy` or `append` to a new slice:

```go
c := make([]int, len(a))
copy(c, a)
```

### Maps

```go
ages := map[string]int{"Ana": 30}
ages["Bo"] = 25
delete(ages, "Ana")

v, ok := ages["Zed"] // v is 0, ok is false
if !ok {
	fmt.Println("not found")
}
```

> [!WARNING]
> Writing to a `nil` map causes a panic. Always create a map with `make` or a literal.
> Map order is random. Never depend on it.

---

## Exercises

### Exercise 1: Reverse and Evens

- `Reverse(nums []int) []int` returns a **new** reversed slice. The input must not change.
- `Evens(nums []int) []int` returns only the even numbers, in the same order.

**Starter** `ch05/ex01/slices.go`

```go
package slices

func Reverse(nums []int) []int {
	return nil // TODO
}

func Evens(nums []int) []int {
	return nil // TODO
}
```

**Test** `ch05/ex01/slices_test.go`

```go
package slices

import (
	"reflect"
	"testing"
)

func TestReverse(t *testing.T) {
	in := []int{1, 2, 3, 4}
	got := Reverse(in)
	if want := []int{4, 3, 2, 1}; !reflect.DeepEqual(got, want) {
		t.Errorf("Reverse = %v, want %v", got, want)
	}
	if !reflect.DeepEqual(in, []int{1, 2, 3, 4}) {
		t.Errorf("Reverse changed its input: %v", in)
	}
}

func TestReverseEmpty(t *testing.T) {
	if got := Reverse(nil); len(got) != 0 {
		t.Errorf("Reverse(nil) = %v, want empty", got)
	}
}

func TestEvens(t *testing.T) {
	got := Evens([]int{1, 2, 3, 4, 5, 6, -2})
	want := []int{2, 4, 6, -2}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Evens = %v, want %v", got, want)
	}
	if got := Evens([]int{1, 3}); len(got) != 0 {
		t.Errorf("Evens of odd numbers = %v, want empty", got)
	}
}
```

<details>
<summary>Solution</summary>

```go
package slices

func Reverse(nums []int) []int {
	out := make([]int, len(nums))
	for i, n := range nums {
		out[len(nums)-1-i] = n
	}
	return out
}

func Evens(nums []int) []int {
	var out []int
	for _, n := range nums {
		if n%2 == 0 {
			out = append(out, n)
		}
	}
	return out
}
```

</details>

### Exercise 2: Clone and Remove

- `Clone(s []int) []int` returns an independent copy. Changing the copy must not change the original.
- `Remove(s []int, i int) []int` returns a new slice without the element at index `i`. The order of the other items stays the same. If `i` is out of range, return a copy of `s` unchanged. The input must not change.

**Starter** `ch05/ex02/copy.go`

```go
package copyx

func Clone(s []int) []int {
	return nil // TODO
}

func Remove(s []int, i int) []int {
	return nil // TODO
}
```

**Test** `ch05/ex02/copy_test.go`

```go
package copyx

import (
	"reflect"
	"testing"
)

func TestClone(t *testing.T) {
	orig := []int{1, 2, 3}
	c := Clone(orig)
	if !reflect.DeepEqual(c, orig) {
		t.Fatalf("Clone = %v, want %v", c, orig)
	}
	c[0] = 99
	if orig[0] != 1 {
		t.Errorf("changing the clone changed the original: %v", orig)
	}
}

func TestRemove(t *testing.T) {
	orig := []int{10, 20, 30, 40}

	got := Remove(orig, 1)
	if want := []int{10, 30, 40}; !reflect.DeepEqual(got, want) {
		t.Errorf("Remove(orig, 1) = %v, want %v", got, want)
	}
	if !reflect.DeepEqual(orig, []int{10, 20, 30, 40}) {
		t.Errorf("Remove changed its input: %v", orig)
	}

	got = Remove(orig, 3)
	if want := []int{10, 20, 30}; !reflect.DeepEqual(got, want) {
		t.Errorf("Remove(orig, 3) = %v, want %v", got, want)
	}
}

func TestRemoveOutOfRange(t *testing.T) {
	orig := []int{1, 2, 3}
	for _, i := range []int{-1, 3, 100} {
		got := Remove(orig, i)
		if !reflect.DeepEqual(got, orig) {
			t.Errorf("Remove(orig, %d) = %v, want %v", i, got, orig)
		}
	}
}
```

<details>
<summary>Solution</summary>

```go
package copyx

func Clone(s []int) []int {
	out := make([]int, len(s))
	copy(out, s)
	return out
}

func Remove(s []int, i int) []int {
	if i < 0 || i >= len(s) {
		return Clone(s)
	}
	out := make([]int, 0, len(s)-1)
	out = append(out, s[:i]...)
	out = append(out, s[i+1:]...)
	return out
}
```

</details>

### Exercise 3: Word count and Unique

- `WordCount(words []string) map[string]int` counts how many times each word appears.
- `Unique(nums []int) []int` removes duplicates and keeps the order of first appearance.

**Starter** `ch05/ex03/maps.go`

```go
package maps

func WordCount(words []string) map[string]int {
	return nil // TODO
}

func Unique(nums []int) []int {
	return nil // TODO
}
```

**Test** `ch05/ex03/maps_test.go`

```go
package maps

import (
	"reflect"
	"testing"
)

func TestWordCount(t *testing.T) {
	got := WordCount([]string{"go", "is", "fun", "go", "go", "is"})
	want := map[string]int{"go": 3, "is": 2, "fun": 1}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("WordCount = %v, want %v", got, want)
	}
}

func TestWordCountEmpty(t *testing.T) {
	got := WordCount(nil)
	if got == nil {
		t.Fatal("WordCount(nil) should return an empty map, not nil")
	}
	if len(got) != 0 {
		t.Errorf("len = %d, want 0", len(got))
	}
}

func TestUnique(t *testing.T) {
	got := Unique([]int{3, 1, 3, 2, 1, 4})
	want := []int{3, 1, 2, 4}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Unique = %v, want %v", got, want)
	}
}
```

<details>
<summary>Solution</summary>

```go
package maps

func WordCount(words []string) map[string]int {
	counts := make(map[string]int)
	for _, w := range words {
		counts[w]++
	}
	return counts
}

func Unique(nums []int) []int {
	seen := make(map[int]bool)
	var out []int
	for _, n := range nums {
		if !seen[n] {
			seen[n] = true
			out = append(out, n)
		}
	}
	return out
}
```

</details>

### Exercise 4: Group and sort keys

- `GroupByLength(words []string) map[int][]string` groups words by their length. Words in a group keep the input order.
- `SortedKeys(m map[string]int) []string` returns the keys in alphabetical order. Map order is random, so you must sort.

> [!TIP]
> Use `sort.Strings(keys)` from the `sort` package.

**Starter** `ch05/ex04/group.go`

```go
package group

func GroupByLength(words []string) map[int][]string {
	return nil // TODO
}

func SortedKeys(m map[string]int) []string {
	return nil // TODO
}
```

**Test** `ch05/ex04/group_test.go`

```go
package group

import (
	"reflect"
	"testing"
)

func TestGroupByLength(t *testing.T) {
	got := GroupByLength([]string{"go", "is", "fun", "and", "simple", "ok"})
	want := map[int][]string{
		2: {"go", "is", "ok"},
		3: {"fun", "and"},
		6: {"simple"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("GroupByLength = %v, want %v", got, want)
	}
}

func TestSortedKeys(t *testing.T) {
	m := map[string]int{"pear": 1, "apple": 2, "mango": 3, "banana": 4}
	got := SortedKeys(m)
	want := []string{"apple", "banana", "mango", "pear"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("SortedKeys = %v, want %v", got, want)
	}
}

func TestSortedKeysEmpty(t *testing.T) {
	if got := SortedKeys(map[string]int{}); len(got) != 0 {
		t.Errorf("SortedKeys(empty) = %v, want empty", got)
	}
}
```

<details>
<summary>Solution</summary>

```go
package group

import "sort"

func GroupByLength(words []string) map[int][]string {
	groups := make(map[int][]string)
	for _, w := range words {
		groups[len(w)] = append(groups[len(w)], w)
	}
	return groups
}

func SortedKeys(m map[string]int) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
```

</details>

## Checkpoint

- [ ] I know a sub slice shares memory with the original
- [ ] I can make an independent copy of a slice
- [ ] I can check if a key exists in a map with `v, ok := m[key]`
