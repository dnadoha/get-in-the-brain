# Chapter 22: Graceful Shutdown and Rate Limiting

## What you will learn

- Setting safe timeouts on `http.Server`
- Shutting down without cutting off users
- Reacting to `Ctrl+C` and container stop signals
- A token bucket rate limiter
- Per client rate limiting as middleware

## Key ideas

### http.ListenAndServe is not enough for production

The shortcut `http.ListenAndServe(addr, h)` has **no timeouts**. A slow or broken client can hold a
connection open forever and use up your server. Create your own `http.Server`:

```go
srv := &http.Server{
	Addr:              ":8080",
	Handler:           mux,
	ReadHeaderTimeout: 5 * time.Second,   // time to read request headers
	ReadTimeout:       10 * time.Second,  // time to read the whole request
	WriteTimeout:      30 * time.Second,  // time to write the response
	IdleTimeout:       2 * time.Minute,   // keep-alive connections
}
```

### Graceful shutdown

When you deploy a new version, the old process receives a signal (`SIGTERM`). If it just exits, requests in
progress are cut. A graceful shutdown:

1. Stops accepting **new** connections.
1. Waits for requests in progress to finish (up to a time limit).
1. Then exits.

```go
ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
defer stop()

go srv.ListenAndServe()

<-ctx.Done() // wait for Ctrl+C or SIGTERM

shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
srv.Shutdown(shutdownCtx)
```

`ListenAndServe` and `Serve` return `http.ErrServerClosed` after `Shutdown`. That is **not** a real error.

### Rate limiting: the token bucket

- A bucket holds up to `burst` tokens. It starts full.
- Every request takes one token. No token, no request.
- Tokens come back at `rate` tokens per second, but never above `burst`.

```text
rate = 2/s, burst = 5
  5 quick requests: allowed (bucket empty)
  6th request:      denied (429)
  after 0.5 s:      1 token is back, one more request is allowed
```

To test time based code, **pass the clock in** (`now func() time.Time`) instead of calling `time.Now()` inside.
Then your tests can move time forward without `time.Sleep`.

---

## Exercises

### Exercise 1: A server with timeouts

Write `NewServer(addr string, h http.Handler) *http.Server` with these values:

| Field | Value |
|-------|-------|
| `Addr` | `addr` |
| `Handler` | `h` |
| `ReadHeaderTimeout` | 5 seconds |
| `ReadTimeout` | 10 seconds |
| `WriteTimeout` | 30 seconds |
| `IdleTimeout` | 2 minutes |

**Starter** `ch22/ex01/newserver.go`

```go
package server

import "net/http"

func NewServer(addr string, h http.Handler) *http.Server {
	return &http.Server{} // TODO
}
```

**Test** `ch22/ex01/newserver_test.go`

```go
package server

import (
	"net/http"
	"testing"
	"time"
)

func TestNewServer(t *testing.T) {
	h := http.NewServeMux()
	srv := NewServer(":9090", h)

	if srv.Addr != ":9090" {
		t.Errorf("Addr = %q, want :9090", srv.Addr)
	}
	if srv.Handler != http.Handler(h) {
		t.Error("Handler was not set")
	}
	if srv.ReadHeaderTimeout != 5*time.Second {
		t.Errorf("ReadHeaderTimeout = %v, want 5s", srv.ReadHeaderTimeout)
	}
	if srv.ReadTimeout != 10*time.Second {
		t.Errorf("ReadTimeout = %v, want 10s", srv.ReadTimeout)
	}
	if srv.WriteTimeout != 30*time.Second {
		t.Errorf("WriteTimeout = %v, want 30s", srv.WriteTimeout)
	}
	if srv.IdleTimeout != 2*time.Minute {
		t.Errorf("IdleTimeout = %v, want 2m", srv.IdleTimeout)
	}
}
```

<details>
<summary>Solution</summary>

```go
package server

import (
	"net/http"
	"time"
)

func NewServer(addr string, h http.Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           h,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       2 * time.Minute,
	}
}
```

</details>

### Exercise 2: Run and shut down gracefully

Write `Run(ctx context.Context, srv *http.Server, ln net.Listener, timeout time.Duration) error`.

- Start `srv.Serve(ln)` in a goroutine.
- When `ctx` is cancelled, call `srv.Shutdown` with a new context that has the given `timeout`.
- A clean shutdown returns `nil`. Requests in progress are allowed to finish.
- If `Shutdown` runs out of time, return an error that wraps the cause (`shutdown: %w`).
- If `Serve` fails by itself (for example a closed listener), return that error right away.
- `http.ErrServerClosed` is not an error.

> [!NOTE]
> The listener is passed in so the test can use a free port (`127.0.0.1:0`).
> In `main` you would create it with `net.Listen("tcp", ":8080")`.

**Starter** `ch22/ex02/run.go`

```go
package server

import (
	"context"
	"net"
	"net/http"
	"time"
)

func Run(ctx context.Context, srv *http.Server, ln net.Listener, timeout time.Duration) error {
	return nil // TODO
}
```

**Test** `ch22/ex02/run_test.go`

```go
package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"
	"time"
)

func newClient() *http.Client {
	return &http.Client{
		Timeout:   2 * time.Second,
		Transport: &http.Transport{DisableKeepAlives: true},
	}
}

func listen(t *testing.T) (net.Listener, string) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	return ln, "http://" + ln.Addr().String()
}

func waitUntilUp(t *testing.T, url string) {
	t.Helper()
	client := newClient()
	deadline := time.Now().Add(3 * time.Second)
	for {
		resp, err := client.Get(url + "/ping")
		if err == nil {
			resp.Body.Close()
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("server did not start: %v", err)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func waitFor(t *testing.T, done <-chan error) error {
	t.Helper()
	select {
	case err := <-done:
		return err
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not return in time")
		return nil
	}
}

func TestRunStopsWhenContextIsCancelled(t *testing.T) {
	ln, url := listen(t)
	srv := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "pong")
	})}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- Run(ctx, srv, ln, 2*time.Second) }()

	waitUntilUp(t, url)
	cancel()

	if err := waitFor(t, done); err != nil {
		t.Errorf("Run returned %v, want nil after a clean shutdown", err)
	}
	if _, err := newClient().Get(url + "/ping"); err == nil {
		t.Error("the server is still accepting requests after shutdown")
	}
}

func TestRunFinishesRequestsInProgress(t *testing.T) {
	ln, url := listen(t)
	started := make(chan struct{}, 1)
	srv := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/slow" {
			started <- struct{}{}
			time.Sleep(300 * time.Millisecond)
			fmt.Fprint(w, "done")
			return
		}
		fmt.Fprint(w, "pong")
	})}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- Run(ctx, srv, ln, 3*time.Second) }()
	waitUntilUp(t, url)

	type result struct {
		body string
		err  error
	}
	results := make(chan result, 1)
	go func() {
		resp, err := newClient().Get(url + "/slow")
		if err != nil {
			results <- result{err: err}
			return
		}
		defer resp.Body.Close()
		b, err := io.ReadAll(resp.Body)
		results <- result{body: string(b), err: err}
	}()

	<-started // the slow request is now running
	cancel()  // ask for shutdown while it runs

	res := <-results
	if res.err != nil || res.body != "done" {
		t.Errorf("slow request got (%q, %v), want (done, nil): it must be allowed to finish", res.body, res.err)
	}
	if err := waitFor(t, done); err != nil {
		t.Errorf("Run returned %v, want nil", err)
	}
}

func TestRunShutdownTimeout(t *testing.T) {
	ln, url := listen(t)
	started := make(chan struct{}, 1)
	release := make(chan struct{})
	srv := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/hang" {
			started <- struct{}{}
			<-release
			return
		}
		fmt.Fprint(w, "pong")
	})}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- Run(ctx, srv, ln, 100*time.Millisecond) }()
	waitUntilUp(t, url)

	go newClient().Get(url + "/hang")
	<-started
	cancel()

	err := waitFor(t, done)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("err = %v, want it to wrap context.DeadlineExceeded", err)
	}

	close(release)
	srv.Close()
}

func TestRunReturnsServeError(t *testing.T) {
	ln, _ := listen(t)
	ln.Close() // Serve will fail at once
	srv := &http.Server{}

	done := make(chan error, 1)
	go func() { done <- Run(context.Background(), srv, ln, time.Second) }()

	if err := waitFor(t, done); err == nil {
		t.Error("expected the Serve error to be returned")
	}
}
```

<details>
<summary>Solution</summary>

```go
package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"
)

func Run(ctx context.Context, srv *http.Server, ln net.Listener, timeout time.Duration) error {
	errCh := make(chan error, 1)
	go func() { errCh <- srv.Serve(ln) }()

	select {
	case err := <-errCh:
		// Serve stopped on its own
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}
	<-errCh // Serve returns http.ErrServerClosed right after Shutdown
	return nil
}
```

Use a **new** context for `Shutdown`. The old one is already cancelled, so `Shutdown` would give up at once.

</details>

### Exercise 3: A token bucket limiter

Write a `Limiter`.

- `NewLimiter(rate float64, burst int, now func() time.Time) *Limiter`. The bucket starts **full** (`burst` tokens).
- `Allow() bool` takes one token if there is at least one and returns `true`. Otherwise it returns `false`.
- Before each decision, add `elapsed seconds * rate` tokens, but never go above `burst`.
- If the clock goes backwards, add nothing.
- It must be safe for many goroutines.

**Starter** `ch22/ex03/limiter.go`

```go
package ratelimit

import "time"

type Limiter struct {
	// TODO
}

func NewLimiter(rate float64, burst int, now func() time.Time) *Limiter {
	return &Limiter{} // TODO
}

func (l *Limiter) Allow() bool {
	return false // TODO
}
```

**Test** `ch22/ex03/limiter_test.go`

```go
package ratelimit

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type fakeClock struct{ t time.Time }

func newClock() *fakeClock                { return &fakeClock{t: time.Unix(1_700_000_000, 0)} }
func (c *fakeClock) Now() time.Time       { return c.t }
func (c *fakeClock) Advance(d time.Duration) { c.t = c.t.Add(d) }

func count(l *Limiter, tries int) int {
	allowed := 0
	for i := 0; i < tries; i++ {
		if l.Allow() {
			allowed++
		}
	}
	return allowed
}

func TestBurst(t *testing.T) {
	clock := newClock()
	l := NewLimiter(1, 3, clock.Now)

	if got := count(l, 10); got != 3 {
		t.Errorf("allowed %d requests at once, want 3 (the burst)", got)
	}
}

func TestRefill(t *testing.T) {
	clock := newClock()
	l := NewLimiter(1, 3, clock.Now)
	count(l, 3) // empty the bucket

	clock.Advance(time.Second)
	if got := count(l, 5); got != 1 {
		t.Errorf("after 1s allowed %d, want 1", got)
	}

	clock.Advance(2 * time.Second)
	if got := count(l, 5); got != 2 {
		t.Errorf("after 2 more seconds allowed %d, want 2", got)
	}
}

func TestRefillIsCappedAtBurst(t *testing.T) {
	clock := newClock()
	l := NewLimiter(1, 3, clock.Now)
	count(l, 3)

	clock.Advance(time.Hour)
	if got := count(l, 10); got != 3 {
		t.Errorf("after a long wait allowed %d, want 3 (never more than burst)", got)
	}
}

func TestFractionalTokens(t *testing.T) {
	clock := newClock()
	l := NewLimiter(2, 1, clock.Now) // 2 tokens per second
	count(l, 1)

	clock.Advance(250 * time.Millisecond) // half a token
	if l.Allow() {
		t.Error("half a token must not allow a request")
	}
	clock.Advance(250 * time.Millisecond) // now one full token
	if !l.Allow() {
		t.Error("one full token should allow a request")
	}
}

func TestClockGoingBackwards(t *testing.T) {
	clock := newClock()
	l := NewLimiter(1, 1, clock.Now)
	count(l, 1)

	clock.Advance(-time.Hour)
	if l.Allow() {
		t.Error("a clock that moves backwards must not give tokens")
	}
}

func TestConcurrent(t *testing.T) {
	clock := newClock() // time never moves, so no refill
	l := NewLimiter(1, 100, clock.Now)

	var allowed atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < 300; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if l.Allow() {
				allowed.Add(1)
			}
		}()
	}
	wg.Wait()

	if got := allowed.Load(); got != 100 {
		t.Errorf("allowed %d, want exactly 100", got)
	}
}
```

<details>
<summary>Solution</summary>

```go
package ratelimit

import (
	"math"
	"sync"
	"time"
)

type Limiter struct {
	mu     sync.Mutex
	rate   float64
	burst  float64
	tokens float64
	last   time.Time
	now    func() time.Time
}

func NewLimiter(rate float64, burst int, now func() time.Time) *Limiter {
	return &Limiter{
		rate:   rate,
		burst:  float64(burst),
		tokens: float64(burst),
		last:   now(),
		now:    now,
	}
}

func (l *Limiter) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	t := l.now()
	if elapsed := t.Sub(l.last).Seconds(); elapsed > 0 {
		l.tokens = math.Min(l.burst, l.tokens+elapsed*l.rate)
		l.last = t
	}

	if l.tokens >= 1 {
		l.tokens--
		return true
	}
	return false
}
```

</details>

### Exercise 4: Rate limit per client

This exercise is in the **same package** as exercise 3.

1. Create `ch22/ex04/` and copy your `limiter.go` from exercise 3 into it (package `ratelimit`).
1. Add `middleware.go` below.

`RateLimit(rate float64, burst int, now func() time.Time) func(http.Handler) http.Handler`:

- Every client gets its **own** `Limiter`. A client is the host part of `r.RemoteAddr` (use `net.SplitHostPort`; if that fails, use the whole string).
- If the limiter allows the request, call the next handler.
- Otherwise respond `429`, header `Retry-After: 1`, header `Content-Type: application/json`, and body `{"error":"too many requests"}`. Do **not** call the next handler.
- The map of limiters must be safe for many goroutines.

> [!NOTE]
> A real service must also delete limiters of clients that have not been seen for a while, or the map grows forever.
> That is left out here to keep the exercise small.

**Starter** `ch22/ex04/middleware.go`

```go
package ratelimit

import (
	"net/http"
	"time"
)

func RateLimit(rate float64, burst int, now func() time.Time) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return next // TODO
	}
}
```

**Test** `ch22/ex04/middleware_test.go`

```go
package ratelimit

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type testClock struct{ t time.Time }

func (c *testClock) Now() time.Time          { return c.t }
func (c *testClock) Advance(d time.Duration) { c.t = c.t.Add(d) }

func setup() (http.Handler, *testClock) {
	clock := &testClock{t: time.Unix(1_700_000_000, 0)}
	h := RateLimit(1, 2, clock.Now)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "ok")
	}))
	return h, clock
}

func get(h http.Handler, remoteAddr string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = remoteAddr
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestAllowsUpToBurstThenBlocks(t *testing.T) {
	h, _ := setup()

	for i := 1; i <= 2; i++ {
		if rec := get(h, "1.2.3.4:5000"); rec.Code != http.StatusOK {
			t.Fatalf("request %d: status = %d, want 200", i, rec.Code)
		}
	}

	rec := get(h, "1.2.3.4:5000")
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("third request: status = %d, want 429", rec.Code)
	}
	if got := rec.Header().Get("Retry-After"); got != "1" {
		t.Errorf("Retry-After = %q, want 1", got)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil || body["error"] != "too many requests" {
		t.Errorf("body = %q, want {\"error\":\"too many requests\"}", rec.Body.String())
	}
}

func TestClientsAreIndependent(t *testing.T) {
	h, _ := setup()

	get(h, "1.1.1.1:1000")
	get(h, "1.1.1.1:1000")
	if rec := get(h, "1.1.1.1:1000"); rec.Code != http.StatusTooManyRequests {
		t.Fatalf("first client should be blocked, got %d", rec.Code)
	}

	if rec := get(h, "2.2.2.2:1000"); rec.Code != http.StatusOK {
		t.Errorf("a different client was blocked: status = %d", rec.Code)
	}
}

func TestSameHostDifferentPortsShareALimit(t *testing.T) {
	h, _ := setup()

	get(h, "3.3.3.3:1111")
	get(h, "3.3.3.3:2222")
	if rec := get(h, "3.3.3.3:3333"); rec.Code != http.StatusTooManyRequests {
		t.Errorf("the port must not count as a different client, status = %d", rec.Code)
	}
}

func TestRecoversOverTime(t *testing.T) {
	h, clock := setup()

	get(h, "4.4.4.4:1")
	get(h, "4.4.4.4:1")
	if rec := get(h, "4.4.4.4:1"); rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected a block, got %d", rec.Code)
	}

	clock.Advance(time.Second)
	if rec := get(h, "4.4.4.4:1"); rec.Code != http.StatusOK {
		t.Errorf("after 1s: status = %d, want 200", rec.Code)
	}
}

func TestBlockedRequestDoesNotReachHandler(t *testing.T) {
	clock := &testClock{t: time.Unix(1_700_000_000, 0)}
	calls := 0
	h := RateLimit(1, 1, clock.Now)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
	}))

	get(h, "5.5.5.5:1")
	get(h, "5.5.5.5:1")
	get(h, "5.5.5.5:1")
	if calls != 1 {
		t.Errorf("handler ran %d times, want 1", calls)
	}
}

func TestRemoteAddrWithoutPort(t *testing.T) {
	h, _ := setup()
	get(h, "6.6.6.6")
	get(h, "6.6.6.6")
	if rec := get(h, "6.6.6.6"); rec.Code != http.StatusTooManyRequests {
		t.Errorf("status = %d, want 429", rec.Code)
	}
}
```

<details>
<summary>Solution (middleware.go)</summary>

```go
package ratelimit

import (
	"encoding/json"
	"net"
	"net/http"
	"sync"
	"time"
)

func RateLimit(rate float64, burst int, now func() time.Time) func(http.Handler) http.Handler {
	var mu sync.Mutex
	limiters := make(map[string]*Limiter)

	limiterFor := func(key string) *Limiter {
		mu.Lock()
		defer mu.Unlock()
		l, ok := limiters[key]
		if !ok {
			l = NewLimiter(rate, burst, now)
			limiters[key] = l
		}
		return l
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			host, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				host = r.RemoteAddr
			}

			if !limiterFor(host).Allow() {
				w.Header().Set("Retry-After", "1")
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				json.NewEncoder(w).Encode(map[string]string{"error": "too many requests"})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
```

</details>

### Put it together in main

```go
func main() {
	cfg, err := config.Load(os.Getenv)
	if err != nil {
		log.Fatal(err)
	}

	handler := Chain(router,
		RequestID(newID),
		RequestLogger(logger),
		Recover(),
		RateLimit(10, 20, time.Now),
	)

	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
	if err != nil {
		log.Fatal(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	srv := NewServer(ln.Addr().String(), handler)
	if err := Run(ctx, srv, ln, 10*time.Second); err != nil {
		log.Fatal(err)
	}
}
```

## Checkpoint

- [ ] I never ship a server without timeouts
- [ ] I can shut a server down gracefully and know why `Shutdown` needs a fresh context
- [ ] I can explain a token bucket with rate and burst
- [ ] I pass the clock into time based code so I can test it
