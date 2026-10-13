# Chapter 13: Goroutines and Channels

## What you will learn

- Starting goroutines with `go`
- Waiting for them with `sync.WaitGroup`
- Sending and receiving on channels
- Closing channels and ranging over them
- Pipelines, worker pools and fan in

## Key ideas

### Goroutines

A goroutine is a very light thread managed by Go. Start one with the `go` keyword.

```go
go doWork()
go func() {
	fmt.Println("hello from a goroutine")
}()
```

`main` does not wait for goroutines. When `main` ends, everything stops.
Use a `sync.WaitGroup` to wait.

```go
var wg sync.WaitGroup
for i := 0; i < 3; i++ {
	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Println("worker", i)
	}()
}
wg.Wait()
```

> [!NOTE]
> Since Go 1.22 each loop iteration has its own copy of `i`, so the code above is safe.
> In older versions you had to pass `i` as an argument.

### Channels

A channel passes values between goroutines.

```go
ch := make(chan int)       // unbuffered: send waits for a receiver
buf := make(chan int, 10)  // buffered: send waits only when full

go func() { ch <- 42 }()
v := <-ch
```

### Closing and ranging

The **sender** closes the channel when it has nothing more to send. A receiver can then use `range`.

```go
go func() {
	defer close(ch)
	for i := 0; i < 3; i++ {
		ch <- i
	}
}()

for v := range ch {
	fmt.Println(v)
}
```

### Direction types

```go
func producer() <-chan int   // you can only receive from the result
func consume(in <-chan int)  // function can only receive
func send(out chan<- int)    // function can only send
```

### Common mistakes

| Mistake | What happens |
|---------|--------------|
| Receiving from a channel nobody sends to | Deadlock, the program freezes |
| Sending on a closed channel | Panic |
| Forgetting `wg.Add` before `go` | `Wait` may return too early |
| Writing the same variable from many goroutines | Data race, run `go test -race` |

---

## Exercises

> [!TIP]
> Run these tests with the race detector: `go test -race ./ch13/...`

### Exercise 1: Parallel squares

Write `ParallelSquares(nums []int) []int`.

- Start one goroutine per number.
- The result keeps the **same order** as the input.
- Use a `WaitGroup`. Each goroutine writes only to its own index, so no lock is needed.

**Starter** `ch13/ex01/squares.go`

```go
package squares

func ParallelSquares(nums []int) []int {
	return nil // TODO
}
```

**Test** `ch13/ex01/squares_test.go`

```go
package squares

import (
	"reflect"
	"testing"
)

func TestParallelSquares(t *testing.T) {
	got := ParallelSquares([]int{1, 2, 3, 4})
	want := []int{1, 4, 9, 16}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestParallelSquaresEmpty(t *testing.T) {
	if got := ParallelSquares(nil); len(got) != 0 {
		t.Errorf("got %v, want empty", got)
	}
}

func TestParallelSquaresMany(t *testing.T) {
	in := make([]int, 1000)
	want := make([]int, 1000)
	for i := range in {
		in[i] = i
		want[i] = i * i
	}
	if got := ParallelSquares(in); !reflect.DeepEqual(got, want) {
		t.Error("result is wrong or in the wrong order")
	}
}
```

<details>
<summary>Solution</summary>

```go
package squares

import "sync"

func ParallelSquares(nums []int) []int {
	out := make([]int, len(nums))
	var wg sync.WaitGroup
	for i, n := range nums {
		wg.Add(1)
		go func() {
			defer wg.Done()
			out[i] = n * n
		}()
	}
	wg.Wait()
	return out
}
```

</details>

### Exercise 2: Generator and collector

- `Generate(n int) <-chan int` sends `1, 2, ..., n` and then closes the channel. It must run in its own goroutine and return at once.
- `Collect(ch <-chan int) []int` reads until the channel is closed and returns all values.

**Starter** `ch13/ex02/gen.go`

```go
package gen

func Generate(n int) <-chan int {
	return nil // TODO
}

func Collect(ch <-chan int) []int {
	return nil // TODO
}
```

**Test** `ch13/ex02/gen_test.go`

```go
package gen

import (
	"reflect"
	"testing"
	"time"
)

func TestGenerateAndCollect(t *testing.T) {
	got := Collect(Generate(5))
	want := []int{1, 2, 3, 4, 5}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestGenerateZero(t *testing.T) {
	if got := Collect(Generate(0)); len(got) != 0 {
		t.Errorf("got %v, want empty", got)
	}
}

func TestGenerateReturnsImmediately(t *testing.T) {
	done := make(chan struct{})
	go func() {
		Generate(100) // nobody reads, it must still return
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Generate blocked: start the sending in a goroutine")
	}
}
```

<details>
<summary>Solution</summary>

```go
package gen

func Generate(n int) <-chan int {
	ch := make(chan int)
	go func() {
		defer close(ch)
		for i := 1; i <= n; i++ {
			ch <- i
		}
	}()
	return ch
}

func Collect(ch <-chan int) []int {
	var out []int
	for v := range ch {
		out = append(out, v)
	}
	return out
}
```

</details>

### Exercise 3: A pipeline

A pipeline connects stages with channels.

- `Square(in <-chan int) <-chan int` sends the square of each value, then closes its output when `in` is closed.
- `Sum(in <-chan int) int` adds all values until the channel closes.

Then the test connects `Generate`, `Square` and `Sum`.

**Starter** `ch13/ex03/pipeline.go`

```go
package pipeline

func Square(in <-chan int) <-chan int {
	return nil // TODO
}

func Sum(in <-chan int) int {
	return 0 // TODO
}
```

**Test** `ch13/ex03/pipeline_test.go`

```go
package pipeline

import "testing"

func fromSlice(vals ...int) <-chan int {
	ch := make(chan int)
	go func() {
		defer close(ch)
		for _, v := range vals {
			ch <- v
		}
	}()
	return ch
}

func TestSquare(t *testing.T) {
	out := Square(fromSlice(1, 2, 3))
	var got []int
	for v := range out {
		got = append(got, v)
	}
	want := []int{1, 4, 9}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("got %v, want %v", got, want)
			break
		}
	}
}

func TestSum(t *testing.T) {
	if got := Sum(fromSlice(1, 2, 3, 4)); got != 10 {
		t.Errorf("Sum = %d, want 10", got)
	}
	if got := Sum(fromSlice()); got != 0 {
		t.Errorf("Sum of nothing = %d, want 0", got)
	}
}

func TestWholePipeline(t *testing.T) {
	// 1 + 4 + 9 + 16 + 25 = 55
	if got := Sum(Square(fromSlice(1, 2, 3, 4, 5))); got != 55 {
		t.Errorf("pipeline result = %d, want 55", got)
	}
}
```

<details>
<summary>Solution</summary>

```go
package pipeline

func Square(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for v := range in {
			out <- v * v
		}
	}()
	return out
}

func Sum(in <-chan int) int {
	total := 0
	for v := range in {
		total += v
	}
	return total
}
```

</details>

### Exercise 4: Worker pool

Write `WorkerPool(jobs []int, workers int, f func(int) int) []int`.

- Start exactly `workers` goroutines (if `workers < 1`, use 1).
- Each worker takes jobs from a channel and calls `f`.
- The result at index `i` is `f(jobs[i])`. The order matches the input.
- At no time may more than `workers` calls to `f` run at once.

> [!TIP]
> Send the **index** of each job on the channel. A worker then writes `results[i] = f(jobs[i])`.

**Starter** `ch13/ex04/pool.go`

```go
package pool

func WorkerPool(jobs []int, workers int, f func(int) int) []int {
	return nil // TODO
}
```

**Test** `ch13/ex04/pool_test.go`

```go
package pool

import (
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestWorkerPoolResults(t *testing.T) {
	jobs := []int{1, 2, 3, 4, 5, 6}
	got := WorkerPool(jobs, 3, func(n int) int { return n * 10 })
	want := []int{10, 20, 30, 40, 50, 60}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestWorkerPoolEmpty(t *testing.T) {
	if got := WorkerPool(nil, 3, func(n int) int { return n }); len(got) != 0 {
		t.Errorf("got %v, want empty", got)
	}
}

func TestWorkerPoolZeroWorkers(t *testing.T) {
	got := WorkerPool([]int{1, 2}, 0, func(n int) int { return n + 1 })
	if !reflect.DeepEqual(got, []int{2, 3}) {
		t.Errorf("got %v, want [2 3]", got)
	}
}

func maxConcurrent(t *testing.T, workers, jobs int) int {
	t.Helper()
	var mu sync.Mutex
	current, peak := 0, 0

	in := make([]int, jobs)
	WorkerPool(in, workers, func(n int) int {
		mu.Lock()
		current++
		if current > peak {
			peak = current
		}
		mu.Unlock()

		time.Sleep(10 * time.Millisecond)

		mu.Lock()
		current--
		mu.Unlock()
		return n
	})
	return peak
}

func TestWorkerPoolLimitsConcurrency(t *testing.T) {
	if peak := maxConcurrent(t, 3, 9); peak > 3 {
		t.Errorf("peak concurrency = %d, want at most 3", peak)
	}
	if peak := maxConcurrent(t, 1, 4); peak != 1 {
		t.Errorf("with 1 worker, peak concurrency = %d, want 1", peak)
	}
}

func TestWorkerPoolRunsInParallel(t *testing.T) {
	if peak := maxConcurrent(t, 3, 9); peak < 2 {
		t.Errorf("peak concurrency = %d, expected the workers to run in parallel", peak)
	}
}
```

<details>
<summary>Solution</summary>

```go
package pool

import "sync"

func WorkerPool(jobs []int, workers int, f func(int) int) []int {
	if workers < 1 {
		workers = 1
	}
	results := make([]int, len(jobs))
	indexes := make(chan int)

	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range indexes {
				results[i] = f(jobs[i])
			}
		}()
	}

	for i := range jobs {
		indexes <- i
	}
	close(indexes)
	wg.Wait()
	return results
}
```

</details>

### Exercise 5: Fan in

Write `Merge(chs ...<-chan int) <-chan int`. It reads from all input channels at the same time and
sends every value to one output channel. The output closes when **all** inputs are closed.

The order of values in the output does not matter.

**Starter** `ch13/ex05/merge.go`

```go
package merge

func Merge(chs ...<-chan int) <-chan int {
	return nil // TODO
}
```

**Test** `ch13/ex05/merge_test.go`

```go
package merge

import (
	"reflect"
	"sort"
	"testing"
)

func fromSlice(vals ...int) <-chan int {
	ch := make(chan int)
	go func() {
		defer close(ch)
		for _, v := range vals {
			ch <- v
		}
	}()
	return ch
}

func drain(ch <-chan int) []int {
	var out []int
	for v := range ch {
		out = append(out, v)
	}
	sort.Ints(out)
	return out
}

func TestMerge(t *testing.T) {
	out := Merge(fromSlice(1, 4, 7), fromSlice(2, 5), fromSlice(3, 6, 8, 9))
	got := drain(out)
	want := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestMergeSingle(t *testing.T) {
	got := drain(Merge(fromSlice(3, 1, 2)))
	if !reflect.DeepEqual(got, []int{1, 2, 3}) {
		t.Errorf("got %v, want [1 2 3]", got)
	}
}

func TestMergeNothing(t *testing.T) {
	if got := drain(Merge()); len(got) != 0 {
		t.Errorf("got %v, want empty (and the channel must close)", got)
	}
}
```

<details>
<summary>Solution</summary>

```go
package merge

import "sync"

func Merge(chs ...<-chan int) <-chan int {
	out := make(chan int)
	var wg sync.WaitGroup

	for _, c := range chs {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for v := range c {
				out <- v
			}
		}()
	}

	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}
```

</details>

## Checkpoint

- [ ] I can start goroutines and wait for them
- [ ] I know who should close a channel (the sender)
- [ ] I can build a small pipeline and a worker pool
- [ ] I run my concurrent tests with `-race`
