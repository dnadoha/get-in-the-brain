# Chapter 12: Generics

## What you will learn

- Type parameters on functions and types
- Constraints: `any`, `comparable`, `cmp.Ordered`, your own
- When generics help and when they do not

## Key ideas

### A generic function

Instead of writing `MaxInt`, `MaxFloat`, `MaxString`, write one function.

```go
func Max[T cmp.Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}

Max(3, 7)       // T is int
Max("a", "b")   // T is string
Max[float64](1, 2.5)
```

The part in square brackets is the **type parameter list**. `cmp.Ordered` is a constraint:
it allows all types that support `<` and `>`.

### Common constraints

| Constraint | Meaning |
|------------|---------|
| `any` | Any type |
| `comparable` | Types that support `==` (map keys need this) |
| `cmp.Ordered` | Numbers and strings, supports `<` `>` |
| your own | A list of types, like `~int \| ~float64` |

The `~` means "this type, or any type whose underlying type is this". So `~int` also accepts
`type Age int`.

### A generic type

```go
type Pair[K comparable, V any] struct {
	Key   K
	Value V
}

p := Pair[string, int]{"age", 30}
```

Methods on a generic type repeat the type parameters in the receiver:

```go
func (p Pair[K, V]) String() string {
	return fmt.Sprintf("%v=%v", p.Key, p.Value)
}
```

> [!TIP]
> Use generics for data structures and small helpers like `Map` and `Filter`.
> If a plain interface is enough, use the interface. Do not make everything generic.

---

## Exercises

### Exercise 1: Map, Filter, Reduce

- `Map[T, U any](s []T, f func(T) U) []U`
- `Filter[T any](s []T, keep func(T) bool) []T`
- `Reduce[T, A any](s []T, initial A, f func(A, T) A) A`

None of them may change the input slice.

**Starter** `ch12/ex01/fp.go`

```go
package fp

func Map[T, U any](s []T, f func(T) U) []U {
	return nil // TODO
}

func Filter[T any](s []T, keep func(T) bool) []T {
	return nil // TODO
}

func Reduce[T, A any](s []T, initial A, f func(A, T) A) A {
	var zero A
	return zero // TODO
}
```

**Test** `ch12/ex01/fp_test.go`

```go
package fp

import (
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestMap(t *testing.T) {
	got := Map([]int{1, 2, 3}, func(n int) string { return strconv.Itoa(n * 10) })
	want := []string{"10", "20", "30"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Map = %v, want %v", got, want)
	}
	if got := Map([]int{}, func(n int) int { return n }); len(got) != 0 {
		t.Errorf("Map on empty = %v, want empty", got)
	}
}

func TestFilter(t *testing.T) {
	words := []string{"go", "rust", "c", "python"}
	got := Filter(words, func(s string) bool { return len(s) > 2 })
	want := []string{"rust", "python"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Filter = %v, want %v", got, want)
	}
	if !reflect.DeepEqual(words, []string{"go", "rust", "c", "python"}) {
		t.Errorf("Filter changed its input: %v", words)
	}
}

func TestReduce(t *testing.T) {
	sum := Reduce([]int{1, 2, 3, 4}, 0, func(acc, n int) int { return acc + n })
	if sum != 10 {
		t.Errorf("sum = %d, want 10", sum)
	}

	// the result type can be different from the element type
	joined := Reduce([]int{1, 2, 3}, "", func(acc string, n int) string {
		return acc + strconv.Itoa(n)
	})
	if joined != "123" {
		t.Errorf("joined = %q, want %q", joined, "123")
	}

	upper := Reduce([]string{"a", "b"}, "", func(acc string, s string) string {
		return acc + strings.ToUpper(s)
	})
	if upper != "AB" {
		t.Errorf("upper = %q, want AB", upper)
	}
}
```

<details>
<summary>Solution</summary>

```go
package fp

func Map[T, U any](s []T, f func(T) U) []U {
	out := make([]U, 0, len(s))
	for _, v := range s {
		out = append(out, f(v))
	}
	return out
}

func Filter[T any](s []T, keep func(T) bool) []T {
	var out []T
	for _, v := range s {
		if keep(v) {
			out = append(out, v)
		}
	}
	return out
}

func Reduce[T, A any](s []T, initial A, f func(A, T) A) A {
	acc := initial
	for _, v := range s {
		acc = f(acc, v)
	}
	return acc
}
```

</details>

### Exercise 2: Constraints

- Define `Number` as a constraint for `~int`, `~int64` and `~float64`.
- `Sum[T Number](nums []T) T` adds the numbers.
- `MaxOf[T cmp.Ordered](first T, rest ...T) T` returns the biggest value.

**Starter** `ch12/ex02/constraints.go`

```go
package constraints

import "cmp"

// TODO: type Number interface { ... }

func Sum[T any](nums []T) T { // TODO: change `any` to Number
	var zero T
	return zero
}

func MaxOf[T cmp.Ordered](first T, rest ...T) T {
	return first // TODO
}
```

**Test** `ch12/ex02/constraints_test.go`

```go
package constraints

import "testing"

type Score int

func TestSum(t *testing.T) {
	if got := Sum([]int{1, 2, 3}); got != 6 {
		t.Errorf("Sum ints = %d, want 6", got)
	}
	if got := Sum([]float64{1.5, 2.5}); got != 4 {
		t.Errorf("Sum floats = %v, want 4", got)
	}
	if got := Sum([]Score{10, 20}); got != 30 {
		t.Errorf("Sum Score = %d, want 30 (did you use ~int?)", got)
	}
	if got := Sum([]int{}); got != 0 {
		t.Errorf("Sum empty = %d, want 0", got)
	}
}

func TestMaxOf(t *testing.T) {
	if got := MaxOf(3, 9, 4); got != 9 {
		t.Errorf("MaxOf ints = %d, want 9", got)
	}
	if got := MaxOf("pear", "apple", "zebra", "mango"); got != "zebra" {
		t.Errorf("MaxOf strings = %q, want zebra", got)
	}
	if got := MaxOf(2.5); got != 2.5 {
		t.Errorf("MaxOf single = %v, want 2.5", got)
	}
}
```

<details>
<summary>Solution</summary>

```go
package constraints

import "cmp"

type Number interface {
	~int | ~int64 | ~float64
}

func Sum[T Number](nums []T) T {
	var total T
	for _, n := range nums {
		total += n
	}
	return total
}

func MaxOf[T cmp.Ordered](first T, rest ...T) T {
	best := first
	for _, v := range rest {
		if v > best {
			best = v
		}
	}
	return best
}
```

</details>

### Exercise 3: Generic Stack

Turn the stack from chapter 07 into `Stack[T any]`. The zero value must be ready to use.

- `Push(v T)`
- `Pop() (T, bool)` returns the zero value of `T` and `false` when empty
- `Peek() (T, bool)`
- `Len() int`

**Starter** `ch12/ex03/stack.go`

```go
package stack

type Stack[T any] struct {
	// TODO
}

// TODO: Push, Pop, Peek, Len
```

**Test** `ch12/ex03/stack_test.go`

```go
package stack

import "testing"

func TestIntStack(t *testing.T) {
	var s Stack[int]
	s.Push(1)
	s.Push(2)
	if s.Len() != 2 {
		t.Fatalf("Len = %d, want 2", s.Len())
	}
	if v, ok := s.Peek(); !ok || v != 2 {
		t.Errorf("Peek = (%d, %v), want (2, true)", v, ok)
	}
	if v, ok := s.Pop(); !ok || v != 2 {
		t.Errorf("Pop = (%d, %v), want (2, true)", v, ok)
	}
	if v, ok := s.Pop(); !ok || v != 1 {
		t.Errorf("Pop = (%d, %v), want (1, true)", v, ok)
	}
	if v, ok := s.Pop(); ok || v != 0 {
		t.Errorf("Pop on empty = (%d, %v), want (0, false)", v, ok)
	}
}

func TestStringStack(t *testing.T) {
	var s Stack[string]
	s.Push("a")
	s.Push("b")
	if v, _ := s.Pop(); v != "b" {
		t.Errorf("Pop = %q, want b", v)
	}
	s.Pop()
	if v, ok := s.Pop(); ok || v != "" {
		t.Errorf("Pop on empty = (%q, %v), want (\"\", false)", v, ok)
	}
}

func TestStructStack(t *testing.T) {
	type point struct{ X, Y int }
	var s Stack[point]
	s.Push(point{1, 2})
	if v, ok := s.Peek(); !ok || v != (point{1, 2}) {
		t.Errorf("Peek = (%v, %v), want ({1 2}, true)", v, ok)
	}
}
```

<details>
<summary>Solution</summary>

```go
package stack

type Stack[T any] struct {
	items []T
}

func (s *Stack[T]) Push(v T) {
	s.items = append(s.items, v)
}

func (s *Stack[T]) Pop() (T, bool) {
	var zero T
	if len(s.items) == 0 {
		return zero, false
	}
	last := len(s.items) - 1
	v := s.items[last]
	s.items = s.items[:last]
	return v, true
}

func (s *Stack[T]) Peek() (T, bool) {
	var zero T
	if len(s.items) == 0 {
		return zero, false
	}
	return s.items[len(s.items)-1], true
}

func (s *Stack[T]) Len() int {
	return len(s.items)
}
```

</details>

### Exercise 4: Generic Set

Build `Set[T comparable]` on top of a map.

- `NewSet[T comparable](items ...T) Set[T]`
- `(s Set[T]) Add(v T)`
- `(s Set[T]) Remove(v T)`
- `(s Set[T]) Has(v T) bool`
- `(s Set[T]) Len() int`
- `Union[T comparable](a, b Set[T]) Set[T]` returns a **new** set. Neither `a` nor `b` may change.

> [!TIP]
> `map[T]struct{}` is a common way to build a set. `struct{}` uses no memory.

**Starter** `ch12/ex04/set.go`

```go
package set

type Set[T comparable] map[T]struct{}

func NewSet[T comparable](items ...T) Set[T] {
	return nil // TODO
}

// TODO: Add, Remove, Has, Len

func Union[T comparable](a, b Set[T]) Set[T] {
	return nil // TODO
}
```

**Test** `ch12/ex04/set_test.go`

```go
package set

import "testing"

func TestSetBasics(t *testing.T) {
	s := NewSet(1, 2, 3, 2, 1)
	if s.Len() != 3 {
		t.Errorf("Len = %d, want 3", s.Len())
	}
	if !s.Has(2) || s.Has(9) {
		t.Error("Has gave a wrong answer")
	}

	s.Add(9)
	if !s.Has(9) || s.Len() != 4 {
		t.Errorf("after Add: Has(9)=%v Len=%d", s.Has(9), s.Len())
	}

	s.Remove(1)
	if s.Has(1) || s.Len() != 3 {
		t.Errorf("after Remove: Has(1)=%v Len=%d", s.Has(1), s.Len())
	}

	s.Remove(12345) // removing a missing value is fine
}

func TestNewSetEmpty(t *testing.T) {
	s := NewSet[string]()
	if s == nil {
		t.Fatal("NewSet() must return a usable (non-nil) set")
	}
	s.Add("a") // would panic on a nil map
	if !s.Has("a") {
		t.Error("Add did not work on an empty set")
	}
}

func TestUnion(t *testing.T) {
	a := NewSet("x", "y")
	b := NewSet("y", "z")
	u := Union(a, b)

	for _, v := range []string{"x", "y", "z"} {
		if !u.Has(v) {
			t.Errorf("union is missing %q", v)
		}
	}
	if u.Len() != 3 {
		t.Errorf("union Len = %d, want 3", u.Len())
	}
	if a.Len() != 2 || b.Len() != 2 {
		t.Errorf("Union changed its inputs: a=%d b=%d", a.Len(), b.Len())
	}
}
```

<details>
<summary>Solution</summary>

```go
package set

type Set[T comparable] map[T]struct{}

func NewSet[T comparable](items ...T) Set[T] {
	s := make(Set[T], len(items))
	for _, v := range items {
		s[v] = struct{}{}
	}
	return s
}

func (s Set[T]) Add(v T)      { s[v] = struct{}{} }
func (s Set[T]) Remove(v T)   { delete(s, v) }
func (s Set[T]) Len() int     { return len(s) }
func (s Set[T]) Has(v T) bool { _, ok := s[v]; return ok }

func Union[T comparable](a, b Set[T]) Set[T] {
	out := make(Set[T], len(a)+len(b))
	for v := range a {
		out[v] = struct{}{}
	}
	for v := range b {
		out[v] = struct{}{}
	}
	return out
}
```

</details>

## Checkpoint

- [ ] I can write a generic function with a constraint
- [ ] I know what `comparable` and `cmp.Ordered` allow
- [ ] I can write a generic type and its methods
