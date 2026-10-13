# Chapter 14: Sync, Select and Context

## What you will learn

- Protecting shared data with `sync.Mutex`
- Running something only once with `sync.Once`
- Waiting on many channels with `select`
- Timeouts and cancellation with `context`

## Key ideas

### Mutex

When many goroutines touch the same variable, you need a lock.

```go
type Counter struct {
	mu sync.Mutex
	n  int
}

func (c *Counter) Inc() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.n++
}
```

Always `defer` the unlock. Never copy a struct that contains a mutex.

> [!TIP]
> Run `go test -race` to find data races. It will print the exact lines.

### sync.Once

Runs a function exactly one time, even when many goroutines call it at once.

```go
var once sync.Once
once.Do(func() { fmt.Println("only once") })
```

### select

`select` waits for the first channel operation that is ready.

```go
select {
case v := <-ch:
	fmt.Println("got", v)
case <-time.After(time.Second):
	fmt.Println("timeout")
default:
	fmt.Println("nothing ready, do not wait")
}
```

Without `default`, `select` blocks until one case is ready.

### context

A `context.Context` carries a cancel signal and a deadline through your code.
Every function that can block or take long should accept one as the **first parameter**.

```go
ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
defer cancel() // always call cancel

select {
case <-ctx.Done():
	fmt.Println(ctx.Err()) // context.DeadlineExceeded or context.Canceled
case result := <-work:
	fmt.Println(result)
}
```

| Function | Use |
|----------|-----|
| `context.Background()` | The root context, in `main` and tests |
| `context.WithCancel(parent)` | You decide when to cancel |
| `context.WithTimeout(parent, d)` | Cancel after a duration |
| `ctx.Done()` | Channel that closes on cancel |
| `ctx.Err()` | Why it was cancelled |

HTTP servers give every request a context. If the client disconnects, it is cancelled.

---

## Exercises

> [!TIP]
> Run with the race detector: `go test -race ./ch14/...`

### Exercise 1: Safe counter

Build `SafeCounter` that counts words from many goroutines.

- `NewSafeCounter() *SafeCounter`
- `Inc(key string)` adds 1 to the count of `key`
- `Value(key string) int` returns the count (0 if the key is unknown)

**Starter** `ch14/ex01/counter.go`

```go
package counter

type SafeCounter struct {
	// TODO: a mutex and a map
}

func NewSafeCounter() *SafeCounter {
	return &SafeCounter{}
}

// TODO: Inc, Value
```

**Test** `ch14/ex01/counter_test.go`

```go
package counter

import (
	"sync"
	"testing"
)

func TestSafeCounter(t *testing.T) {
	c := NewSafeCounter()
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				c.Inc("a")
			}
			c.Inc("b")
		}()
	}
	wg.Wait()

	if got := c.Value("a"); got != 10000 {
		t.Errorf("Value(a) = %d, want 10000", got)
	}
	if got := c.Value("b"); got != 100 {
		t.Errorf("Value(b) = %d, want 100", got)
	}
	if got := c.Value("missing"); got != 0 {
		t.Errorf("Value(missing) = %d, want 0", got)
	}
}
```

<details>
<summary>Solution</summary>

```go
package counter

import "sync"

type SafeCounter struct {
	mu     sync.Mutex
	counts map[string]int
}

func NewSafeCounter() *SafeCounter {
	return &SafeCounter{counts: make(map[string]int)}
}

func (c *SafeCounter) Inc(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.counts[key]++
}

func (c *SafeCounter) Value(key string) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.counts[key]
}
```

</details>

### Exercise 2: Timeout and non blocking receive

- `ReceiveWithTimeout(ch <-chan string, d time.Duration) (string, error)` waits for a value for at most `d`. If nothing arrives in time, return `ErrTimeout`.
- `TryReceive(ch <-chan int) (int, bool)` returns a value if one is ready **right now**, otherwise `0, false`. It must never wait.

**Starter** `ch14/ex02/recv.go`

```go
package recv

import (
	"errors"
	"time"
)

var ErrTimeout = errors.New("timeout")

func ReceiveWithTimeout(ch <-chan string, d time.Duration) (string, error) {
	return "", nil // TODO
}

func TryReceive(ch <-chan int) (int, bool) {
	return 0, false // TODO
}
```

**Test** `ch14/ex02/recv_test.go`

```go
package recv

import (
	"errors"
	"testing"
	"time"
)

func TestReceiveValueArrives(t *testing.T) {
	ch := make(chan string)
	go func() {
		time.Sleep(10 * time.Millisecond)
		ch <- "hello"
	}()
	got, err := ReceiveWithTimeout(ch, time.Second)
	if err != nil || got != "hello" {
		t.Errorf("got (%q, %v), want (hello, nil)", got, err)
	}
}

func TestReceiveTimeout(t *testing.T) {
	ch := make(chan string) // nobody sends
	start := time.Now()
	_, err := ReceiveWithTimeout(ch, 30*time.Millisecond)
	if !errors.Is(err, ErrTimeout) {
		t.Errorf("err = %v, want ErrTimeout", err)
	}
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Errorf("took %v, should give up after about 30ms", elapsed)
	}
}

func TestTryReceive(t *testing.T) {
	ch := make(chan int, 1)

	if v, ok := TryReceive(ch); ok || v != 0 {
		t.Errorf("empty channel: got (%d, %v), want (0, false)", v, ok)
	}

	ch <- 5
	if v, ok := TryReceive(ch); !ok || v != 5 {
		t.Errorf("with a value: got (%d, %v), want (5, true)", v, ok)
	}

	if _, ok := TryReceive(ch); ok {
		t.Error("the value was already taken, expected false")
	}
}
```

<details>
<summary>Solution</summary>

```go
package recv

import (
	"errors"
	"time"
)

var ErrTimeout = errors.New("timeout")

func ReceiveWithTimeout(ch <-chan string, d time.Duration) (string, error) {
	select {
	case v := <-ch:
		return v, nil
	case <-time.After(d):
		return "", ErrTimeout
	}
}

func TryReceive(ch <-chan int) (int, bool) {
	select {
	case v := <-ch:
		return v, true
	default:
		return 0, false
	}
}
```

</details>

### Exercise 3: Respect the context

Write `SlowOperation(ctx context.Context, d time.Duration) error`. It pretends to work for `d`.

- If it finishes, return `nil`.
- If the context ends first, stop right away and return `ctx.Err()`.

**Starter** `ch14/ex03/slow.go`

```go
package slow

import (
	"context"
	"time"
)

func SlowOperation(ctx context.Context, d time.Duration) error {
	return nil // TODO
}
```

**Test** `ch14/ex03/slow_test.go`

```go
package slow

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestSlowOperationCompletes(t *testing.T) {
	if err := SlowOperation(context.Background(), 10*time.Millisecond); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSlowOperationCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	err := SlowOperation(ctx, 5*time.Second)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("err = %v, want context.Canceled", err)
	}
	if time.Since(start) > time.Second {
		t.Error("the operation did not stop when the context was cancelled")
	}
}

func TestSlowOperationTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	err := SlowOperation(ctx, 5*time.Second)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("err = %v, want context.DeadlineExceeded", err)
	}
}

func TestSlowOperationAlreadyCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := SlowOperation(ctx, 5*time.Second); !errors.Is(err, context.Canceled) {
		t.Errorf("err = %v, want context.Canceled", err)
	}
}
```

<details>
<summary>Solution</summary>

```go
package slow

import (
	"context"
	"time"
)

func SlowOperation(ctx context.Context, d time.Duration) error {
	select {
	case <-time.After(d):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
```

</details>

### Exercise 4: Stop a producer

Write `CountUntilCancelled(ctx context.Context, out chan<- int)`.

- It sends `1, 2, 3, ...` on `out`, one number at a time.
- When `ctx` is cancelled, the function returns. This must also work while it is **waiting to send**.
- It does not close `out`.

> [!TIP]
> Use `select` with both `out <- i` and `<-ctx.Done()`.

**Starter** `ch14/ex04/count.go`

```go
package count

import "context"

func CountUntilCancelled(ctx context.Context, out chan<- int) {
	// TODO
}
```

**Test** `ch14/ex04/count_test.go`

```go
package count

import (
	"context"
	"testing"
	"time"
)

func receive(t *testing.T, ch <-chan int) int {
	t.Helper()
	select {
	case v := <-ch:
		return v
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for a number")
		return 0
	}
}

func TestCountUntilCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	out := make(chan int)

	done := make(chan struct{})
	go func() {
		CountUntilCancelled(ctx, out)
		close(done)
	}()

	for want := 1; want <= 3; want++ {
		if got := receive(t, out); got != want {
			t.Errorf("got %d, want %d", got, want)
		}
	}

	cancel()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("CountUntilCancelled did not return after cancel")
	}
}

func TestCountStopsEvenIfNobodyReads(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		CountUntilCancelled(ctx, make(chan int)) // no reader at all
		close(done)
	}()

	time.Sleep(20 * time.Millisecond)
	cancel()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("blocked on send: use select with ctx.Done()")
	}
}
```

<details>
<summary>Solution</summary>

```go
package count

import "context"

func CountUntilCancelled(ctx context.Context, out chan<- int) {
	for i := 1; ; i++ {
		select {
		case out <- i:
		case <-ctx.Done():
			return
		}
	}
}
```

</details>

### Exercise 5: Lazy value with sync.Once

Build `Lazy`. It calls `load` only when `Get` is called the first time. Later calls return the same value.
Even if 50 goroutines call `Get` at once, `load` must run exactly once.

**Starter** `ch14/ex05/lazy.go`

```go
package lazy

type Lazy struct {
	// TODO
}

func NewLazy(load func() int) *Lazy {
	return &Lazy{} // TODO
}

func (l *Lazy) Get() int {
	return 0 // TODO
}
```

**Test** `ch14/ex05/lazy_test.go`

```go
package lazy

import (
	"sync"
	"sync/atomic"
	"testing"
)

func TestLazyNotCalledEarly(t *testing.T) {
	var calls atomic.Int32
	NewLazy(func() int {
		calls.Add(1)
		return 1
	})
	if calls.Load() != 0 {
		t.Error("load must not run until Get is called")
	}
}

func TestLazyRunsOnce(t *testing.T) {
	var calls atomic.Int32
	l := NewLazy(func() int {
		calls.Add(1)
		return 42
	})

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if got := l.Get(); got != 42 {
				t.Errorf("Get = %d, want 42", got)
			}
		}()
	}
	wg.Wait()

	if got := calls.Load(); got != 1 {
		t.Errorf("load ran %d times, want 1", got)
	}
}
```

<details>
<summary>Solution</summary>

```go
package lazy

import "sync"

type Lazy struct {
	once  sync.Once
	load  func() int
	value int
}

func NewLazy(load func() int) *Lazy {
	return &Lazy{load: load}
}

func (l *Lazy) Get() int {
	l.once.Do(func() {
		l.value = l.load()
	})
	return l.value
}
```

</details>

## Checkpoint

- [ ] I protect shared data with a mutex and `defer Unlock`
- [ ] I can add a timeout with `select` and `time.After`
- [ ] I pass `context.Context` as the first parameter and stop when `ctx.Done()` fires
