# Chapter 18: Middleware

## What you will learn

- What middleware is and why it exists
- Wrapping an `http.Handler`
- Capturing the status code by wrapping `ResponseWriter`
- Recovering from panics
- Authentication checks
- Chaining middleware in the right order
- Passing values with `context`

## Key ideas

### Middleware wraps a handler

Middleware is a function that takes a handler and returns a new handler. It can run code
before and after the real handler, or stop the request early.

```go
type Middleware func(http.Handler) http.Handler

func Example() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// before
			next.ServeHTTP(w, r)
			// after
		})
	}
}
```

Typical jobs for middleware: logging, panic recovery, authentication, request ids,
rate limits, CORS, compression.

### Order matters

```text
Chain(h, A, B, C)   means   A( B( C( h ) ) )

request  ->  A  ->  B  ->  C  ->  handler
response <-  A  <-  B  <-  C  <-
```

A common order is: request id, logging, recover, auth.

### Wrapping ResponseWriter

`http.ResponseWriter` does not tell you which status was written. Wrap it to find out.

```go
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}
```

Because `ResponseWriter` is **embedded**, all the other methods still work.
If the handler never calls `WriteHeader`, the status is `200`, so start with `200`.

### Context values

You can attach a value to a request context. Use your own unexported key type, so no other package can clash with it.

```go
type ctxKey struct{}

ctx := context.WithValue(r.Context(), ctxKey{}, "abc")
r = r.WithContext(ctx)

id, _ := r.Context().Value(ctxKey{}).(string)
```

Use context values for request scoped data (request id, current user). Do not use them to pass normal function arguments.

---

## Exercises

### Exercise 1: Logging

Write `Logging(log func(string)) Middleware`. After the inner handler finishes, call `log` with
`"<METHOD> <PATH> <STATUS>"`, for example `"GET /x 404"`.

- If the handler never sets a status, log `200`.
- The response must reach the client unchanged.

**Starter** `ch18/ex01/logging.go`

```go
package logging

import "net/http"

type Middleware func(http.Handler) http.Handler

func Logging(log func(string)) Middleware {
	return func(next http.Handler) http.Handler {
		return next // TODO
	}
}
```

**Test** `ch18/ex01/logging_test.go`

```go
package logging

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLoggingStatus(t *testing.T) {
	var logs []string
	mw := Logging(func(s string) { logs = append(logs, s) })

	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "nothing here")
	}))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/x", nil))

	if len(logs) != 1 || logs[0] != "GET /x 404" {
		t.Errorf("logs = %v, want [GET /x 404]", logs)
	}
	if rec.Code != http.StatusNotFound || rec.Body.String() != "nothing here" {
		t.Errorf("response changed: %d %q", rec.Code, rec.Body.String())
	}
}

func TestLoggingDefaultStatus(t *testing.T) {
	var logs []string
	mw := Logging(func(s string) { logs = append(logs, s) })

	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "hi") // never calls WriteHeader
	}))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/y", nil))

	if len(logs) != 1 || logs[0] != "POST /y 200" {
		t.Errorf("logs = %v, want [POST /y 200]", logs)
	}
	if rec.Body.String() != "hi" {
		t.Errorf("body = %q, want hi", rec.Body.String())
	}
}
```

<details>
<summary>Solution</summary>

```go
package logging

import (
	"fmt"
	"net/http"
)

type Middleware func(http.Handler) http.Handler

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

func Logging(log func(string)) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rec, r)
			log(fmt.Sprintf("%s %s %d", r.Method, r.URL.Path, rec.status))
		})
	}
}
```

</details>

### Exercise 2: Recover from panics

Write `Recover() Middleware`.

- If the inner handler panics, respond with status `500`, header `Content-Type: application/json` and body `{"error":"internal server error"}`.
- If there is no panic, the response is unchanged.
- The server must keep working for the next request.

**Starter** `ch18/ex02/recover.go`

```go
package recoverer

import "net/http"

type Middleware func(http.Handler) http.Handler

func Recover() Middleware {
	return func(next http.Handler) http.Handler {
		return next // TODO
	}
}
```

**Test** `ch18/ex02/recover_test.go`

```go
package recoverer

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRecoverPanic(t *testing.T) {
	h := Recover()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	}))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("body is not JSON: %v (%q)", err, rec.Body.String())
	}
	if body["error"] != "internal server error" {
		t.Errorf("error = %q", body["error"])
	}
}

func TestRecoverNoPanic(t *testing.T) {
	h := Recover()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, "fine")
	}))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusCreated || rec.Body.String() != "fine" {
		t.Errorf("got %d %q, want 201 fine", rec.Code, rec.Body.String())
	}
}

func TestRecoverKeepsWorking(t *testing.T) {
	calls := 0
	h := Recover()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			panic("first request fails")
		}
		fmt.Fprint(w, "ok")
	}))

	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK || rec.Body.String() != "ok" {
		t.Errorf("second request: %d %q, want 200 ok", rec.Code, rec.Body.String())
	}
}
```

<details>
<summary>Solution</summary>

```go
package recoverer

import (
	"encoding/json"
	"net/http"
)

type Middleware func(http.Handler) http.Handler

func Recover() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					// In a real project, log rec and the stack trace here.
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					json.NewEncoder(w).Encode(map[string]string{
						"error": "internal server error",
					})
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
```

</details>

### Exercise 3: API key check

Write `RequireAPIKey(key string) Middleware`.

- The request must have the header `X-API-Key` with exactly the right key.
- If it is missing or wrong, respond `401` with `{"error":"unauthorized"}` and **do not** call the inner handler.
- If `key` is empty, reject every request. An empty configured key must never let people in.

> [!TIP]
> Compare secrets with `subtle.ConstantTimeCompare` from `crypto/subtle`. It takes the same time
> no matter where the strings differ, so attackers cannot learn the key one character at a time.

**Starter** `ch18/ex03/apikey.go`

```go
package apikey

import "net/http"

type Middleware func(http.Handler) http.Handler

func RequireAPIKey(key string) Middleware {
	return func(next http.Handler) http.Handler {
		return next // TODO
	}
}
```

**Test** `ch18/ex03/apikey_test.go`

```go
package apikey

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func serve(key, header string) (*httptest.ResponseRecorder, bool) {
	reached := false
	h := RequireAPIKey(key)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reached = true
		fmt.Fprint(w, "secret data")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if header != "" {
		req.Header.Set("X-API-Key", header)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec, reached
}

func TestCorrectKey(t *testing.T) {
	rec, reached := serve("s3cret", "s3cret")
	if !reached || rec.Code != http.StatusOK || rec.Body.String() != "secret data" {
		t.Errorf("valid key was rejected: reached=%v code=%d", reached, rec.Code)
	}
}

func TestMissingKey(t *testing.T) {
	rec, reached := serve("s3cret", "")
	if reached {
		t.Error("the handler must not run without a key")
	}
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil || body["error"] != "unauthorized" {
		t.Errorf("body = %q, want {\"error\":\"unauthorized\"}", rec.Body.String())
	}
}

func TestWrongKey(t *testing.T) {
	rec, reached := serve("s3cret", "guess")
	if reached || rec.Code != http.StatusUnauthorized {
		t.Errorf("wrong key: reached=%v code=%d, want false 401", reached, rec.Code)
	}
}

func TestEmptyConfiguredKeyRejectsEverything(t *testing.T) {
	for _, header := range []string{"", "anything"} {
		rec, reached := serve("", header)
		if reached || rec.Code != http.StatusUnauthorized {
			t.Errorf("header %q with empty key: reached=%v code=%d, want false 401", header, reached, rec.Code)
		}
	}
}
```

<details>
<summary>Solution</summary>

```go
package apikey

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"
)

type Middleware func(http.Handler) http.Handler

func RequireAPIKey(key string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			got := r.Header.Get("X-API-Key")
			valid := key != "" &&
				subtle.ConstantTimeCompare([]byte(got), []byte(key)) == 1
			if !valid {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
```

</details>

### Exercise 4: Chain

Write `Chain(h http.Handler, mws ...Middleware) http.Handler`.
The **first** middleware in the list is the **outermost**: it runs first on the way in and last on the way out.

**Starter** `ch18/ex04/chain.go`

```go
package chain

import "net/http"

type Middleware func(http.Handler) http.Handler

func Chain(h http.Handler, mws ...Middleware) http.Handler {
	return h // TODO
}
```

**Test** `ch18/ex04/chain_test.go`

```go
package chain

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestChainOrder(t *testing.T) {
	var order []string

	mark := func(name string) Middleware {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, name+" before")
				next.ServeHTTP(w, r)
				order = append(order, name+" after")
			})
		}
	}

	final := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		order = append(order, "handler")
	})

	h := Chain(final, mark("A"), mark("B"), mark("C"))
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	want := []string{
		"A before", "B before", "C before",
		"handler",
		"C after", "B after", "A after",
	}
	if !reflect.DeepEqual(order, want) {
		t.Errorf("order =\n%v\nwant\n%v", order, want)
	}
}

func TestChainWithoutMiddleware(t *testing.T) {
	called := false
	final := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true })

	Chain(final).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
	if !called {
		t.Error("the handler was not called")
	}
}
```

<details>
<summary>Solution</summary>

```go
package chain

import "net/http"

type Middleware func(http.Handler) http.Handler

func Chain(h http.Handler, mws ...Middleware) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}
```

Wrap from the last to the first, so the first ends up on the outside.

</details>

### Exercise 5: Request id in the context

Write `RequestID(newID func() string) Middleware` and `RequestIDFrom(ctx context.Context) string`.

- If the request has an `X-Request-ID` header that is not empty, use it. Otherwise call `newID()`.
- Put the id in the request context and also in the **response** header `X-Request-ID`.
- `RequestIDFrom` returns the id from a context, or `""` if there is none.

**Starter** `ch18/ex05/requestid.go`

```go
package requestid

import (
	"context"
	"net/http"
)

type Middleware func(http.Handler) http.Handler

func RequestID(newID func() string) Middleware {
	return func(next http.Handler) http.Handler {
		return next // TODO
	}
}

func RequestIDFrom(ctx context.Context) string {
	return "" // TODO
}
```

**Test** `ch18/ex05/requestid_test.go`

```go
package requestid

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func run(header string) (seenInHandler string, rec *httptest.ResponseRecorder) {
	h := RequestID(func() string { return "generated-1" })(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			seenInHandler = RequestIDFrom(r.Context())
		}),
	)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if header != "" {
		req.Header.Set("X-Request-ID", header)
	}
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return seenInHandler, rec
}

func TestGeneratedID(t *testing.T) {
	seen, rec := run("")
	if seen != "generated-1" {
		t.Errorf("id in handler = %q, want generated-1", seen)
	}
	if got := rec.Header().Get("X-Request-ID"); got != "generated-1" {
		t.Errorf("response header = %q, want generated-1", got)
	}
}

func TestIncomingIDIsReused(t *testing.T) {
	seen, rec := run("from-client")
	if seen != "from-client" {
		t.Errorf("id in handler = %q, want from-client", seen)
	}
	if got := rec.Header().Get("X-Request-ID"); got != "from-client" {
		t.Errorf("response header = %q, want from-client", got)
	}
}

func TestRequestIDFromEmptyContext(t *testing.T) {
	if got := RequestIDFrom(context.Background()); got != "" {
		t.Errorf("got %q, want an empty string", got)
	}
}
```

<details>
<summary>Solution</summary>

```go
package requestid

import (
	"context"
	"net/http"
)

type Middleware func(http.Handler) http.Handler

type ctxKey struct{}

func RequestID(newID func() string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := r.Header.Get("X-Request-ID")
			if id == "" {
				id = newID()
			}
			w.Header().Set("X-Request-ID", id)

			ctx := context.WithValue(r.Context(), ctxKey{}, id)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequestIDFrom(ctx context.Context) string {
	id, _ := ctx.Value(ctxKey{}).(string)
	return id
}
```

</details>

## Checkpoint

- [ ] I can write a middleware that wraps `http.Handler`
- [ ] I know how to find out the status code a handler wrote
- [ ] I understand the order of a middleware chain
- [ ] I use an unexported key type for context values
