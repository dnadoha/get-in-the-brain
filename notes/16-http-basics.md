# Chapter 16: HTTP Basics

## What you will learn

- How an `http.Handler` works
- Reading query values, path values and JSON bodies
- Writing status codes, headers and JSON
- Routing with the Go 1.22 `ServeMux`
- Testing handlers with `httptest`, with no real server

## Key ideas

### A handler is a function

```go
func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, %s!", r.URL.Query().Get("name"))
}
```

- `r` is the incoming request: method, URL, headers, body.
- `w` is where you write the response.

### A tiny server

```go
func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", health)
	mux.HandleFunc("GET /hello/{name}", hello)

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
```

Try it:

```bash
curl -i localhost:8080/health
```

### Routing patterns (Go 1.22)

| Pattern | Matches |
|---------|---------|
| `GET /health` | only GET requests to `/health` |
| `POST /users` | only POST |
| `GET /users/{id}` | any id, read with `r.PathValue("id")` |
| `/static/` | everything under `/static/` |

The mux answers `404` for unknown paths and `405 Method Not Allowed` for a known path with the wrong method.

### Order matters when writing a response

1. Set headers: `w.Header().Set("Content-Type", "application/json")`
1. Write the status: `w.WriteHeader(http.StatusCreated)`
1. Write the body

If you write the body first, the status becomes `200` and later changes are ignored.

### Common status codes

| Code | Meaning |
|------|---------|
| 200 OK | Success |
| 201 Created | Something was created |
| 204 No Content | Success with no body |
| 400 Bad Request | The client sent something wrong |
| 401 Unauthorized | Not logged in |
| 403 Forbidden | Logged in but not allowed |
| 404 Not Found | It does not exist |
| 409 Conflict | Clash with current state, for example a duplicate |
| 500 Internal Server Error | Your bug |

### Testing without a server

```go
req := httptest.NewRequest(http.MethodGet, "/health", nil)
rec := httptest.NewRecorder()

handler.ServeHTTP(rec, req)

rec.Code           // status code
rec.Body.String()  // body
rec.Header()       // headers
```

---

## Exercises

### Exercise 1: Health check

Write `HealthHandler`. It responds with status `200`, the header `Content-Type: text/plain; charset=utf-8`, and the body `ok`.

**Starter** `ch16/ex01/health.go`

```go
package health

import "net/http"

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	// TODO
}
```

**Test** `ch16/ex01/health_test.go`

```go
package health

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	HealthHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	if got := rec.Body.String(); got != "ok" {
		t.Errorf("body = %q, want %q", got, "ok")
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/plain") {
		t.Errorf("Content-Type = %q, want text/plain", ct)
	}
}
```

<details>
<summary>Solution</summary>

```go
package health

import (
	"fmt"
	"net/http"
)

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "ok")
}
```

</details>

### Exercise 2: Query parameters

Write `GreetHandler`. It reads the query value `name`.

- `/greet?name=Ana` gives `Hello, Ana!`
- `/greet` or `/greet?name=` gives `Hello, World!`

**Starter** `ch16/ex02/greet.go`

```go
package greet

import "net/http"

func GreetHandler(w http.ResponseWriter, r *http.Request) {
	// TODO
}
```

**Test** `ch16/ex02/greet_test.go`

```go
package greet

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGreetHandler(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want string
	}{
		{"with name", "/greet?name=Ana", "Hello, Ana!"},
		{"no name", "/greet", "Hello, World!"},
		{"empty name", "/greet?name=", "Hello, World!"},
		{"spaces", "/greet?name=Ana+Maria", "Hello, Ana Maria!"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.url, nil)
			rec := httptest.NewRecorder()

			GreetHandler(rec, req)

			if rec.Code != http.StatusOK {
				t.Errorf("status = %d, want 200", rec.Code)
			}
			if got := rec.Body.String(); got != tc.want {
				t.Errorf("body = %q, want %q", got, tc.want)
			}
		})
	}
}
```

<details>
<summary>Solution</summary>

```go
package greet

import (
	"fmt"
	"net/http"
)

func GreetHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		name = "World"
	}
	fmt.Fprintf(w, "Hello, %s!", name)
}
```

</details>

### Exercise 3: JSON response helpers

You will use these two helpers in every API you build.

- `WriteJSON(w http.ResponseWriter, status int, v any)` sets `Content-Type: application/json`, writes the status, then writes `v` as JSON.
- `WriteError(w http.ResponseWriter, status int, msg string)` sends `{"error": "<msg>"}` with that status.

**Starter** `ch16/ex03/respond.go`

```go
package respond

import "net/http"

func WriteJSON(w http.ResponseWriter, status int, v any) {
	// TODO
}

func WriteError(w http.ResponseWriter, status int, msg string) {
	// TODO
}
```

**Test** `ch16/ex03/respond_test.go`

```go
package respond

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	WriteJSON(rec, http.StatusCreated, map[string]string{"a": "b"})

	if rec.Code != http.StatusCreated {
		t.Errorf("status = %d, want 201", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
	var got map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("body is not valid JSON: %v (%q)", err, rec.Body.String())
	}
	if got["a"] != "b" {
		t.Errorf("got %v, want a=b", got)
	}
}

func TestWriteJSONStruct(t *testing.T) {
	type item struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	rec := httptest.NewRecorder()
	WriteJSON(rec, http.StatusOK, item{ID: 1, Name: "pen"})

	var got item
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if got.ID != 1 || got.Name != "pen" {
		t.Errorf("got %+v", got)
	}
}

func TestWriteError(t *testing.T) {
	rec := httptest.NewRecorder()
	WriteError(rec, http.StatusNotFound, "user not found")

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
	var got map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if got["error"] != "user not found" {
		t.Errorf("got %v, want error=user not found", got)
	}
}
```

<details>
<summary>Solution</summary>

```go
package respond

import (
	"encoding/json"
	"net/http"
)

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func WriteError(w http.ResponseWriter, status int, msg string) {
	WriteJSON(w, status, map[string]string{"error": msg})
}
```

</details>

### Exercise 4: Read a JSON body

Write `EchoHandler`. It reads `{"message": "..."}` from the request body.

- On success: status `200` and body `{"echo": "<message>"}`.
- If the body is not valid JSON: status `400` and `{"error": "invalid JSON"}`.
- If `message` is empty or missing: status `400` and `{"error": "message is required"}`.

The helpers `WriteJSON` and `WriteError` are already in the starter file.

**Starter** `ch16/ex04/echo.go`

```go
package echo

import (
	"encoding/json"
	"net/http"
)

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func WriteError(w http.ResponseWriter, status int, msg string) {
	WriteJSON(w, status, map[string]string{"error": msg})
}

type echoRequest struct {
	Message string `json:"message"`
}

func EchoHandler(w http.ResponseWriter, r *http.Request) {
	// TODO
}
```

**Test** `ch16/ex04/echo_test.go`

```go
package echo

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func call(body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/echo", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	EchoHandler(rec, req)
	return rec
}

func decode(t *testing.T, rec *httptest.ResponseRecorder) map[string]string {
	t.Helper()
	var m map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &m); err != nil {
		t.Fatalf("response is not valid JSON: %v (%q)", err, rec.Body.String())
	}
	return m
}

func TestEchoOK(t *testing.T) {
	rec := call(`{"message":"hi"}`)
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	if got := decode(t, rec)["echo"]; got != "hi" {
		t.Errorf("echo = %q, want hi", got)
	}
}

func TestEchoInvalidJSON(t *testing.T) {
	rec := call(`{"message":`)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
	if got := decode(t, rec)["error"]; got != "invalid JSON" {
		t.Errorf("error = %q, want %q", got, "invalid JSON")
	}
}

func TestEchoMissingMessage(t *testing.T) {
	for _, body := range []string{`{}`, `{"message":""}`} {
		rec := call(body)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("body %s: status = %d, want 400", body, rec.Code)
		}
		if got := decode(t, rec)["error"]; got != "message is required" {
			t.Errorf("body %s: error = %q, want %q", body, got, "message is required")
		}
	}
}
```

<details>
<summary>Solution</summary>

```go
func EchoHandler(w http.ResponseWriter, r *http.Request) {
	var req echoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if req.Message == "" {
		WriteError(w, http.StatusBadRequest, "message is required")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"echo": req.Message})
}
```

Do not forget the `return` after writing an error. Without it the handler keeps running.

</details>

### Exercise 5: A router

Write `NewRouter() http.Handler` with the Go 1.22 mux.

| Request | Response |
|---------|----------|
| `GET /health` | `200`, body `ok` |
| `GET /hello/{name}` | `200`, body `Hello, <name>!` (read with `r.PathValue("name")`) |
| any other method on these paths | `405` (the mux does this for you) |
| any other path | `404` (the mux does this for you) |

**Starter** `ch16/ex05/router.go`

```go
package router

import "net/http"

func NewRouter() http.Handler {
	mux := http.NewServeMux()
	// TODO: register the two routes
	return mux
}
```

**Test** `ch16/ex05/router_test.go`

```go
package router

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func do(method, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	rec := httptest.NewRecorder()
	NewRouter().ServeHTTP(rec, req)
	return rec
}

func TestHealth(t *testing.T) {
	rec := do(http.MethodGet, "/health")
	if rec.Code != 200 || rec.Body.String() != "ok" {
		t.Errorf("got %d %q, want 200 ok", rec.Code, rec.Body.String())
	}
}

func TestHello(t *testing.T) {
	rec := do(http.MethodGet, "/hello/Ana")
	if rec.Code != 200 || rec.Body.String() != "Hello, Ana!" {
		t.Errorf("got %d %q, want 200 Hello, Ana!", rec.Code, rec.Body.String())
	}
}

func TestWrongMethod(t *testing.T) {
	rec := do(http.MethodPost, "/health")
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", rec.Code)
	}
}

func TestNotFound(t *testing.T) {
	rec := do(http.MethodGet, "/nope")
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}
```

<details>
<summary>Solution</summary>

```go
package router

import (
	"fmt"
	"net/http"
)

func NewRouter() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "ok")
	})

	mux.HandleFunc("GET /hello/{name}", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, %s!", r.PathValue("name"))
	})

	return mux
}
```

</details>

## Checkpoint

- [ ] I can write a handler and test it with `httptest`
- [ ] I set headers before calling `WriteHeader`
- [ ] I use `r.PathValue` and `r.URL.Query()`
- [ ] I `return` after writing an error response
