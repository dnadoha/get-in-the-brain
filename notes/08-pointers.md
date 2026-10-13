# Chapter 08: Pointers

## What you will learn

- What a pointer is: the address of a value
- `&` and `*`
- Why Go passes copies, and how pointers change that
- `nil` pointers and how to stay safe
- A small linked list

## Key ideas

### Address and value

```go
x := 10
p := &x      // p holds the address of x, type *int
fmt.Println(*p) // 10, read through the pointer
*p = 20      // change x through the pointer
fmt.Println(x)  // 20
```

- `&x` means "the address of x"
- `*p` means "the value at address p"

### Go passes copies

```go
func addOne(n int)   { n++ }   // changes a copy
func addOnePtr(n *int) { *n++ } // changes the original

a := 1
addOne(a)     // a is still 1
addOnePtr(&a) // a is 2
```

Structs are copied too. Pass a pointer when a function must change the struct.

### nil

The zero value of a pointer is `nil`. Using `*p` on a nil pointer causes a panic.
Check first when a pointer can be nil.

```go
if p != nil {
	fmt.Println(*p)
}
```

### Pointers to structs

Go lets you skip the star for field access: `p.Name` means `(*p).Name`.

---

## Exercises

### Exercise 1: Swap and Increment

- `Swap(a, b *int)` swaps the two values.
- `Increment(n *int)` adds 1. If `n` is `nil`, do nothing (no panic).

**Starter** `ch08/ex01/ptr.go`

```go
package ptr

func Swap(a, b *int) {
	// TODO
}

func Increment(n *int) {
	// TODO
}
```

**Test** `ch08/ex01/ptr_test.go`

```go
package ptr

import "testing"

func TestSwap(t *testing.T) {
	x, y := 1, 2
	Swap(&x, &y)
	if x != 2 || y != 1 {
		t.Errorf("after Swap: x=%d y=%d, want x=2 y=1", x, y)
	}
}

func TestIncrement(t *testing.T) {
	n := 41
	Increment(&n)
	if n != 42 {
		t.Errorf("n = %d, want 42", n)
	}
}

func TestIncrementNil(t *testing.T) {
	// must not panic
	Increment(nil)
}
```

<details>
<summary>Solution</summary>

```go
package ptr

func Swap(a, b *int) {
	*a, *b = *b, *a
}

func Increment(n *int) {
	if n == nil {
		return
	}
	*n++
}
```

</details>

### Exercise 2: Changing a struct

You have a `Config` struct. Write `SetDefaults(cfg *Config)`.

- If `Host` is empty, set it to `"localhost"`.
- If `Port` is `0`, set it to `8080`.
- Do not touch values that are already set.
- If `cfg` is `nil`, do nothing.

**Starter** `ch08/ex02/config.go`

```go
package config

type Config struct {
	Host string
	Port int
}

func SetDefaults(cfg *Config) {
	// TODO
}
```

**Test** `ch08/ex02/config_test.go`

```go
package config

import "testing"

func TestSetDefaults(t *testing.T) {
	c := Config{}
	SetDefaults(&c)
	if c.Host != "localhost" || c.Port != 8080 {
		t.Errorf("got %+v, want {localhost 8080}", c)
	}
}

func TestSetDefaultsKeepsValues(t *testing.T) {
	c := Config{Host: "example.com", Port: 9000}
	SetDefaults(&c)
	if c.Host != "example.com" || c.Port != 9000 {
		t.Errorf("values were changed: %+v", c)
	}
}

func TestSetDefaultsPartial(t *testing.T) {
	c := Config{Port: 3000}
	SetDefaults(&c)
	if c.Host != "localhost" || c.Port != 3000 {
		t.Errorf("got %+v, want {localhost 3000}", c)
	}
}

func TestSetDefaultsNil(t *testing.T) {
	SetDefaults(nil) // must not panic
}
```

<details>
<summary>Solution</summary>

```go
package config

type Config struct {
	Host string
	Port int
}

func SetDefaults(cfg *Config) {
	if cfg == nil {
		return
	}
	if cfg.Host == "" {
		cfg.Host = "localhost"
	}
	if cfg.Port == 0 {
		cfg.Port = 8080
	}
}
```

</details>

### Exercise 3: Optional values

A pointer can mean "maybe there is a value". This is common in JSON and databases.

- `IntPtr(n int) *int` returns a pointer to a copy of `n`.
- `Deref(p *int, def int) int` returns `*p`, or `def` if `p` is `nil`.

**Starter** `ch08/ex03/optional.go`

```go
package optional

func IntPtr(n int) *int {
	return nil // TODO
}

func Deref(p *int, def int) int {
	return 0 // TODO
}
```

**Test** `ch08/ex03/optional_test.go`

```go
package optional

import "testing"

func TestIntPtr(t *testing.T) {
	p := IntPtr(7)
	if p == nil {
		t.Fatal("IntPtr returned nil")
	}
	if *p != 7 {
		t.Errorf("*p = %d, want 7", *p)
	}
}

func TestIntPtrMakesCopies(t *testing.T) {
	a := IntPtr(1)
	b := IntPtr(1)
	if a == b {
		t.Error("each call should return a new pointer")
	}
}

func TestDeref(t *testing.T) {
	if got := Deref(IntPtr(5), 99); got != 5 {
		t.Errorf("Deref(ptr to 5, 99) = %d, want 5", got)
	}
	if got := Deref(nil, 99); got != 99 {
		t.Errorf("Deref(nil, 99) = %d, want 99", got)
	}
	if got := Deref(IntPtr(0), 99); got != 0 {
		t.Errorf("Deref(ptr to 0, 99) = %d, want 0", got)
	}
}
```

<details>
<summary>Solution</summary>

```go
package optional

func IntPtr(n int) *int {
	return &n
}

func Deref(p *int, def int) int {
	if p == nil {
		return def
	}
	return *p
}
```

</details>

### Exercise 4: Linked list

Build a singly linked list.

- `Push(v int)` adds a value at the **front**
- `Len() int`
- `ToSlice() []int` returns the values from front to back
- `Reverse()` reverses the list in place

The zero value `var l List` must work.

**Starter** `ch08/ex04/list.go`

```go
package list

type Node struct {
	Value int
	Next  *Node
}

type List struct {
	head *Node
	size int
}

// TODO: Push, Len, ToSlice, Reverse (pointer receivers)
```

**Test** `ch08/ex04/list_test.go`

```go
package list

import (
	"reflect"
	"testing"
)

func TestEmptyList(t *testing.T) {
	var l List
	if l.Len() != 0 {
		t.Errorf("Len = %d, want 0", l.Len())
	}
	if got := l.ToSlice(); len(got) != 0 {
		t.Errorf("ToSlice = %v, want empty", got)
	}
	l.Reverse() // must not panic on an empty list
}

func TestPush(t *testing.T) {
	var l List
	l.Push(1)
	l.Push(2)
	l.Push(3)
	if l.Len() != 3 {
		t.Errorf("Len = %d, want 3", l.Len())
	}
	want := []int{3, 2, 1}
	if got := l.ToSlice(); !reflect.DeepEqual(got, want) {
		t.Errorf("ToSlice = %v, want %v", got, want)
	}
}

func TestReverse(t *testing.T) {
	var l List
	for i := 1; i <= 4; i++ {
		l.Push(i)
	}
	// list is 4 3 2 1
	l.Reverse()
	want := []int{1, 2, 3, 4}
	if got := l.ToSlice(); !reflect.DeepEqual(got, want) {
		t.Errorf("after Reverse = %v, want %v", got, want)
	}
	if l.Len() != 4 {
		t.Errorf("Len = %d, want 4", l.Len())
	}
}
```

<details>
<summary>Solution</summary>

```go
package list

type Node struct {
	Value int
	Next  *Node
}

type List struct {
	head *Node
	size int
}

func (l *List) Push(v int) {
	l.head = &Node{Value: v, Next: l.head}
	l.size++
}

func (l *List) Len() int {
	return l.size
}

func (l *List) ToSlice() []int {
	out := make([]int, 0, l.size)
	for n := l.head; n != nil; n = n.Next {
		out = append(out, n.Value)
	}
	return out
}

func (l *List) Reverse() {
	var prev *Node
	cur := l.head
	for cur != nil {
		next := cur.Next
		cur.Next = prev
		prev = cur
		cur = next
	}
	l.head = prev
}
```

</details>

## Checkpoint

- [ ] I can explain `&x` and `*p`
- [ ] I know why a function needs a pointer to change its argument
- [ ] I check for `nil` before using a pointer that may be empty
