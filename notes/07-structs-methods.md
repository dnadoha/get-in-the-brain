# Chapter 07: Structs and Methods

## What you will learn

- Defining structs and creating values
- Methods with value receivers and pointer receivers
- Constructor functions
- Embedding

## Key ideas

### Structs

```go
type User struct {
	Name string
	Age  int
}

u := User{Name: "Ana", Age: 30}
fmt.Println(u.Name)
```

Fields that start with a capital letter are **exported** (visible to other packages).
Fields that start with a small letter are private to the package.

### Methods

A method is a function with a receiver.

```go
// value receiver: works on a copy
func (u User) IsAdult() bool {
	return u.Age >= 18
}

// pointer receiver: can change the original
func (u *User) Birthday() {
	u.Age++
}
```

Use a **pointer receiver** when the method changes the value, or when the struct is big.
Chapter 08 explains pointers in more detail.

### Constructors

Go has no constructors. By convention you write a `NewX` function.

```go
func NewUser(name string, age int) *User {
	return &User{Name: name, Age: age}
}
```

### Embedding

You can put one struct inside another. Its fields and methods are promoted.

```go
type Animal struct{ Name string }

func (a Animal) Speak() string { return a.Name + " makes a sound" }

type Dog struct {
	Animal
	Breed string
}

d := Dog{Animal{"Rex"}, "Lab"}
d.Name    // promoted field
d.Speak() // promoted method
```

---

## Exercises

### Exercise 1: Rectangle

Create a `Rectangle` struct with `Width` and `Height` (both `float64`) and these methods:

- `Area() float64`
- `Perimeter() float64`
- `IsSquare() bool`

**Starter** `ch07/ex01/rect.go`

```go
package rect

type Rectangle struct {
	// TODO: add fields Width and Height
}

// TODO: add the methods Area, Perimeter and IsSquare
```

**Test** `ch07/ex01/rect_test.go`

```go
package rect

import "testing"

func TestArea(t *testing.T) {
	r := Rectangle{Width: 3, Height: 4}
	if got := r.Area(); got != 12 {
		t.Errorf("Area = %v, want 12", got)
	}
}

func TestPerimeter(t *testing.T) {
	r := Rectangle{Width: 3, Height: 4}
	if got := r.Perimeter(); got != 14 {
		t.Errorf("Perimeter = %v, want 14", got)
	}
}

func TestIsSquare(t *testing.T) {
	if !(Rectangle{Width: 5, Height: 5}).IsSquare() {
		t.Error("5x5 should be a square")
	}
	if (Rectangle{Width: 5, Height: 6}).IsSquare() {
		t.Error("5x6 should not be a square")
	}
}
```

<details>
<summary>Solution</summary>

```go
package rect

type Rectangle struct {
	Width  float64
	Height float64
}

func (r Rectangle) Area() float64 {
	return r.Width * r.Height
}

func (r Rectangle) Perimeter() float64 {
	return 2 * (r.Width + r.Height)
}

func (r Rectangle) IsSquare() bool {
	return r.Width == r.Height
}
```

</details>

### Exercise 2: Stack

Build a `Stack` of `int` values. The zero value (`var s Stack`) must be ready to use.

- `Push(v int)` adds a value on top
- `Pop() (int, bool)` removes and returns the top value. If the stack is empty, return `0, false`
- `Peek() (int, bool)` returns the top value without removing it
- `Len() int` returns the number of values

**Starter** `ch07/ex02/stack.go`

```go
package stack

type Stack struct {
	// TODO
}

// TODO: Push, Pop, Peek, Len (use pointer receivers for Push and Pop)
```

**Test** `ch07/ex02/stack_test.go`

```go
package stack

import "testing"

func TestEmptyStack(t *testing.T) {
	var s Stack
	if s.Len() != 0 {
		t.Errorf("Len = %d, want 0", s.Len())
	}
	if v, ok := s.Pop(); ok || v != 0 {
		t.Errorf("Pop on empty = (%d, %v), want (0, false)", v, ok)
	}
	if v, ok := s.Peek(); ok || v != 0 {
		t.Errorf("Peek on empty = (%d, %v), want (0, false)", v, ok)
	}
}

func TestPushPop(t *testing.T) {
	var s Stack
	s.Push(1)
	s.Push(2)
	s.Push(3)

	if s.Len() != 3 {
		t.Fatalf("Len = %d, want 3", s.Len())
	}
	if v, _ := s.Peek(); v != 3 {
		t.Errorf("Peek = %d, want 3", v)
	}
	if s.Len() != 3 {
		t.Errorf("Peek must not remove: Len = %d, want 3", s.Len())
	}

	for _, want := range []int{3, 2, 1} {
		v, ok := s.Pop()
		if !ok || v != want {
			t.Errorf("Pop = (%d, %v), want (%d, true)", v, ok, want)
		}
	}
	if s.Len() != 0 {
		t.Errorf("Len after popping all = %d, want 0", s.Len())
	}
}
```

<details>
<summary>Solution</summary>

```go
package stack

type Stack struct {
	items []int
}

func (s *Stack) Push(v int) {
	s.items = append(s.items, v)
}

func (s *Stack) Pop() (int, bool) {
	if len(s.items) == 0 {
		return 0, false
	}
	last := len(s.items) - 1
	v := s.items[last]
	s.items = s.items[:last]
	return v, true
}

func (s *Stack) Peek() (int, bool) {
	if len(s.items) == 0 {
		return 0, false
	}
	return s.items[len(s.items)-1], true
}

func (s *Stack) Len() int {
	return len(s.items)
}
```

</details>

### Exercise 3: Bank account

Build an `Account` that stores money as whole cents (`int`) to avoid decimal errors.

- `NewAccount(owner string) *Account` starts with a balance of `0`
- `Deposit(cents int) error` fails with an error if `cents <= 0`
- `Withdraw(cents int) error` fails if `cents <= 0` or if there is not enough money. A failed withdraw must not change the balance.
- `Balance() int` returns the balance
- `Owner() string` returns the owner

**Starter** `ch07/ex03/account.go`

```go
package account

type Account struct {
	// TODO
}

func NewAccount(owner string) *Account {
	return nil // TODO
}

// TODO: Deposit, Withdraw, Balance, Owner
```

**Test** `ch07/ex03/account_test.go`

```go
package account

import "testing"

func TestNewAccount(t *testing.T) {
	a := NewAccount("Ana")
	if a.Owner() != "Ana" {
		t.Errorf("Owner = %q, want Ana", a.Owner())
	}
	if a.Balance() != 0 {
		t.Errorf("Balance = %d, want 0", a.Balance())
	}
}

func TestDeposit(t *testing.T) {
	a := NewAccount("Ana")
	if err := a.Deposit(500); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := a.Deposit(250); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Balance() != 750 {
		t.Errorf("Balance = %d, want 750", a.Balance())
	}
}

func TestDepositInvalid(t *testing.T) {
	a := NewAccount("Ana")
	for _, amount := range []int{0, -10} {
		if err := a.Deposit(amount); err == nil {
			t.Errorf("Deposit(%d) should return an error", amount)
		}
	}
	if a.Balance() != 0 {
		t.Errorf("Balance = %d, want 0", a.Balance())
	}
}

func TestWithdraw(t *testing.T) {
	a := NewAccount("Ana")
	a.Deposit(1000)
	if err := a.Withdraw(400); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Balance() != 600 {
		t.Errorf("Balance = %d, want 600", a.Balance())
	}
}

func TestWithdrawFails(t *testing.T) {
	a := NewAccount("Ana")
	a.Deposit(100)

	if err := a.Withdraw(101); err == nil {
		t.Error("withdrawing more than the balance should fail")
	}
	if err := a.Withdraw(0); err == nil {
		t.Error("withdrawing 0 should fail")
	}
	if err := a.Withdraw(-5); err == nil {
		t.Error("withdrawing a negative amount should fail")
	}
	if a.Balance() != 100 {
		t.Errorf("a failed withdraw changed the balance to %d", a.Balance())
	}
}
```

<details>
<summary>Solution</summary>

```go
package account

import "errors"

type Account struct {
	owner   string
	balance int
}

func NewAccount(owner string) *Account {
	return &Account{owner: owner}
}

func (a *Account) Deposit(cents int) error {
	if cents <= 0 {
		return errors.New("deposit must be positive")
	}
	a.balance += cents
	return nil
}

func (a *Account) Withdraw(cents int) error {
	if cents <= 0 {
		return errors.New("withdraw must be positive")
	}
	if cents > a.balance {
		return errors.New("insufficient funds")
	}
	a.balance -= cents
	return nil
}

func (a *Account) Balance() int {
	return a.balance
}

func (a *Account) Owner() string {
	return a.owner
}
```

</details>

### Exercise 4: Embedding

- `Person` has `Name string` and `Age int`, and a method `Intro() string` that returns `"Ana (30)"`.
- `Employee` **embeds** `Person` and adds `Role string`. Its own `Intro()` returns `"Ana (30), Engineer"`.
- `Person.Intro()` must still work through `e.Person.Intro()`.

**Starter** `ch07/ex04/embed.go`

```go
package embed

type Person struct {
	// TODO
}

// TODO: func (p Person) Intro() string

type Employee struct {
	// TODO: embed Person and add Role
}

// TODO: func (e Employee) Intro() string
```

**Test** `ch07/ex04/embed_test.go`

```go
package embed

import "testing"

func TestPersonIntro(t *testing.T) {
	p := Person{Name: "Ana", Age: 30}
	if got := p.Intro(); got != "Ana (30)" {
		t.Errorf("Intro = %q, want %q", got, "Ana (30)")
	}
}

func TestEmployeeIntro(t *testing.T) {
	e := Employee{Person: Person{Name: "Ana", Age: 30}, Role: "Engineer"}
	if got := e.Intro(); got != "Ana (30), Engineer" {
		t.Errorf("Intro = %q, want %q", got, "Ana (30), Engineer")
	}
}

func TestPromotedFieldsAndInnerMethod(t *testing.T) {
	e := Employee{Person: Person{Name: "Bo", Age: 41}, Role: "Manager"}
	if e.Name != "Bo" || e.Age != 41 {
		t.Errorf("promoted fields are wrong: %q %d", e.Name, e.Age)
	}
	if got := e.Person.Intro(); got != "Bo (41)" {
		t.Errorf("Person.Intro = %q, want %q", got, "Bo (41)")
	}
}
```

<details>
<summary>Solution</summary>

```go
package embed

import "fmt"

type Person struct {
	Name string
	Age  int
}

func (p Person) Intro() string {
	return fmt.Sprintf("%s (%d)", p.Name, p.Age)
}

type Employee struct {
	Person
	Role string
}

func (e Employee) Intro() string {
	return e.Person.Intro() + ", " + e.Role
}
```

</details>

## Checkpoint

- [ ] I know when to use a pointer receiver
- [ ] I can write a `NewX` constructor
- [ ] I understand that embedding promotes fields and methods
