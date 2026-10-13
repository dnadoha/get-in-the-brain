# Chapter 17: REST API with CRUD

## What you will learn

- What REST means in practice: resources, methods, status codes
- Building an in-memory store with a mutex
- Writing handlers for create, read, update, delete
- Validating input and returning clear errors
- Filtering with query parameters

## Key ideas

### Resources and methods

A **resource** is a thing your API manages, like a todo. The URL names it, the method says what to do.

| Method | Path | Action | Success code |
|--------|------|--------|--------------|
| `GET` | `/todos` | List all | 200 |
| `POST` | `/todos` | Create one | 201 |
| `GET` | `/todos/{id}` | Read one | 200 |
| `PUT` | `/todos/{id}` | Replace one | 200 |
| `DELETE` | `/todos/{id}` | Remove one | 204 |

### A good handler follows the same steps

1. Read and check the input (path, query, body). Bad input gives `400`.
1. Ask the store to do the work.
1. Turn the result into a response: `404` if missing, `200` or `201` if fine.
1. `return` right after every error response.

### Layers

```text
HTTP handler  ->  Store (data and rules)
reads request     knows nothing about HTTP
writes response   easy to test alone
```

Keeping the store free of `net/http` is the key idea. You test it without any request, and in
chapter 19 you will swap it for a real database.

---

## How this chapter works

You build **one project** in `ch17/todoapi/` in four milestones. The skeleton below already
compiles. All handlers answer `501 Not Implemented` and all store methods are empty.
Each milestone has its own test file. Run only the tests of the milestone you work on:

```bash
go test -run 'TestStore' -v ./ch17/todoapi/       # milestone 1
go test -run 'TestAPI_ListCreate' -v ./ch17/todoapi/  # milestone 2
go test -run 'TestAPI_Item' -v ./ch17/todoapi/    # milestone 3
go test -run 'TestAPI_Filter' -v ./ch17/todoapi/  # milestone 4
```

### Skeleton: files to create first

**Starter** `ch17/todoapi/helpers.go` (finished, do not change)

```go
package todoapi

import (
	"encoding/json"
	"net/http"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
```

**Starter** `ch17/todoapi/store.go`

```go
package todoapi

import (
	"errors"
	"sync"
)

var ErrNotFound = errors.New("todo not found")

type Todo struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

type Store struct {
	mu     sync.Mutex
	nextID int
	todos  map[int]Todo
}

func NewStore() *Store {
	return &Store{nextID: 1, todos: make(map[int]Todo)}
}

// Create adds a todo with the next id and Done set to false.
func (s *Store) Create(title string) Todo {
	return Todo{} // TODO
}

// List returns all todos ordered by ID. It never returns nil.
func (s *Store) List() []Todo {
	return nil // TODO
}

// Get returns the todo with this id, or ErrNotFound.
func (s *Store) Get(id int) (Todo, error) {
	return Todo{}, nil // TODO
}

// Update changes title and done of an existing todo, or returns ErrNotFound.
func (s *Store) Update(id int, title string, done bool) (Todo, error) {
	return Todo{}, nil // TODO
}

// Delete removes a todo, or returns ErrNotFound.
func (s *Store) Delete(id int) error {
	return nil // TODO
}
```

**Starter** `ch17/todoapi/handlers.go`

```go
package todoapi

import "net/http"

type API struct {
	store *Store
}

func NewAPI(store *Store) *API {
	return &API{store: store}
}

func (a *API) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /todos", a.listTodos)
	mux.HandleFunc("POST /todos", a.createTodo)
	mux.HandleFunc("GET /todos/{id}", a.getTodo)
	mux.HandleFunc("PUT /todos/{id}", a.updateTodo)
	mux.HandleFunc("DELETE /todos/{id}", a.deleteTodo)
	return mux
}

func notImplemented(w http.ResponseWriter) {
	writeError(w, http.StatusNotImplemented, "not implemented")
}

func (a *API) listTodos(w http.ResponseWriter, r *http.Request)  { notImplemented(w) }
func (a *API) createTodo(w http.ResponseWriter, r *http.Request) { notImplemented(w) }
func (a *API) getTodo(w http.ResponseWriter, r *http.Request)    { notImplemented(w) }
func (a *API) updateTodo(w http.ResponseWriter, r *http.Request) { notImplemented(w) }
func (a *API) deleteTodo(w http.ResponseWriter, r *http.Request) { notImplemented(w) }
```

**Test** `ch17/todoapi/helpers_test.go` (shared by all milestones, copy this first)

```go
package todoapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newHandler() http.Handler {
	return NewAPI(NewStore()).Routes()
}

func do(h http.Handler, method, path, body string) *httptest.ResponseRecorder {
	var req *http.Request
	if body == "" {
		req = httptest.NewRequest(method, path, nil)
	} else {
		req = httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func decodeTodo(t *testing.T, rec *httptest.ResponseRecorder) Todo {
	t.Helper()
	var todo Todo
	if err := json.Unmarshal(rec.Body.Bytes(), &todo); err != nil {
		t.Fatalf("response is not a valid todo: %v (%q)", err, rec.Body.String())
	}
	return todo
}

func decodeTodos(t *testing.T, rec *httptest.ResponseRecorder) []Todo {
	t.Helper()
	var todos []Todo
	if err := json.Unmarshal(rec.Body.Bytes(), &todos); err != nil {
		t.Fatalf("response is not a valid todo list: %v (%q)", err, rec.Body.String())
	}
	return todos
}

func errorMessage(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()
	var m map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &m); err != nil {
		t.Fatalf("response is not valid JSON: %v (%q)", err, rec.Body.String())
	}
	return m["error"]
}
```

---

## Milestone 1: The store

Fill in the five `Store` methods in `store.go`.

- `Create` gives ids `1, 2, 3, ...`. An id is **never reused**, even after a delete.
- `List` returns todos ordered by id, and an empty (non nil) slice when there are none.
- `Get`, `Update` and `Delete` return `ErrNotFound` for an unknown id.
- The store must be safe to use from many goroutines.

**Test** `ch17/todoapi/store_test.go`

```go
package todoapi

import (
	"errors"
	"sync"
	"testing"
)

func TestStoreCreate(t *testing.T) {
	s := NewStore()
	a := s.Create("Buy milk")
	b := s.Create("Walk dog")

	if a.ID != 1 || b.ID != 2 {
		t.Errorf("ids = %d, %d, want 1, 2", a.ID, b.ID)
	}
	if a.Title != "Buy milk" || a.Done {
		t.Errorf("got %+v, want title Buy milk and Done=false", a)
	}
}

func TestStoreList(t *testing.T) {
	s := NewStore()
	if got := s.List(); got == nil || len(got) != 0 {
		t.Errorf("empty store: List = %#v, want an empty non-nil slice", got)
	}

	s.Create("a")
	s.Create("b")
	s.Create("c")
	got := s.List()
	if len(got) != 3 {
		t.Fatalf("len = %d, want 3", len(got))
	}
	for i, todo := range got {
		if todo.ID != i+1 {
			t.Errorf("List()[%d].ID = %d, want %d (ordered by id)", i, todo.ID, i+1)
		}
	}
}

func TestStoreGet(t *testing.T) {
	s := NewStore()
	created := s.Create("Read")

	got, err := s.Get(created.ID)
	if err != nil || got != created {
		t.Errorf("Get = (%+v, %v), want (%+v, nil)", got, err, created)
	}

	if _, err := s.Get(99); !errors.Is(err, ErrNotFound) {
		t.Errorf("Get(99) err = %v, want ErrNotFound", err)
	}
}

func TestStoreUpdate(t *testing.T) {
	s := NewStore()
	created := s.Create("Old")

	got, err := s.Update(created.ID, "New", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != created.ID || got.Title != "New" || !got.Done {
		t.Errorf("Update returned %+v", got)
	}

	again, _ := s.Get(created.ID)
	if again != got {
		t.Errorf("update was not saved: %+v", again)
	}

	if _, err := s.Update(99, "x", false); !errors.Is(err, ErrNotFound) {
		t.Errorf("Update(99) err = %v, want ErrNotFound", err)
	}
}

func TestStoreDelete(t *testing.T) {
	s := NewStore()
	created := s.Create("Temp")

	if err := s.Delete(created.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := s.Get(created.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("after Delete, Get err = %v, want ErrNotFound", err)
	}
	if err := s.Delete(created.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("second Delete err = %v, want ErrNotFound", err)
	}
}

func TestStoreIDsAreNotReused(t *testing.T) {
	s := NewStore()
	s.Create("one")
	two := s.Create("two")
	s.Delete(two.ID)

	three := s.Create("three")
	if three.ID != 3 {
		t.Errorf("id = %d, want 3 (ids must not be reused)", three.ID)
	}
}

func TestStoreConcurrent(t *testing.T) {
	s := NewStore()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.Create("x")
			s.List()
		}()
	}
	wg.Wait()

	todos := s.List()
	if len(todos) != 50 {
		t.Fatalf("len = %d, want 50", len(todos))
	}
	seen := make(map[int]bool)
	for _, todo := range todos {
		if seen[todo.ID] {
			t.Errorf("duplicate id %d", todo.ID)
		}
		seen[todo.ID] = true
	}
}
```

<details>
<summary>Solution</summary>

```go
package todoapi

import (
	"errors"
	"sort"
	"sync"
)

var ErrNotFound = errors.New("todo not found")

type Todo struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

type Store struct {
	mu     sync.Mutex
	nextID int
	todos  map[int]Todo
}

func NewStore() *Store {
	return &Store{nextID: 1, todos: make(map[int]Todo)}
}

func (s *Store) Create(title string) Todo {
	s.mu.Lock()
	defer s.mu.Unlock()

	todo := Todo{ID: s.nextID, Title: title}
	s.todos[todo.ID] = todo
	s.nextID++
	return todo
}

func (s *Store) List() []Todo {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]Todo, 0, len(s.todos))
	for _, todo := range s.todos {
		out = append(out, todo)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func (s *Store) Get(id int) (Todo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	todo, ok := s.todos[id]
	if !ok {
		return Todo{}, ErrNotFound
	}
	return todo, nil
}

func (s *Store) Update(id int, title string, done bool) (Todo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	todo, ok := s.todos[id]
	if !ok {
		return Todo{}, ErrNotFound
	}
	todo.Title = title
	todo.Done = done
	s.todos[id] = todo
	return todo, nil
}

func (s *Store) Delete(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.todos[id]; !ok {
		return ErrNotFound
	}
	delete(s.todos, id)
	return nil
}
```

</details>

---

## Milestone 2: List and create

Implement `listTodos` and `createTodo`.

- `GET /todos` returns `200` and a JSON **array**. With no todos it must be `[]`, not `null`.
- `POST /todos` with `{"title": "..."}` returns `201` and the new todo.
- The title is trimmed (`"  Read  "` is stored as `"Read"`).
- Invalid JSON gives `400` and `{"error": "invalid JSON"}`.
- An empty or blank title gives `400` and `{"error": "title is required"}`.

**Test** `ch17/todoapi/list_create_test.go`

```go
package todoapi

import (
	"net/http"
	"strings"
	"testing"
)

func TestAPI_ListCreate_EmptyList(t *testing.T) {
	h := newHandler()
	rec := do(h, http.MethodGet, "/todos", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if body := strings.TrimSpace(rec.Body.String()); body != "[]" {
		t.Errorf("body = %q, want []", body)
	}
}

func TestAPI_ListCreate_Create(t *testing.T) {
	h := newHandler()
	rec := do(h, http.MethodPost, "/todos", `{"title":"Buy milk"}`)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
	todo := decodeTodo(t, rec)
	if todo.ID != 1 || todo.Title != "Buy milk" || todo.Done {
		t.Errorf("got %+v, want {1 Buy milk false}", todo)
	}
}

func TestAPI_ListCreate_CreateThenList(t *testing.T) {
	h := newHandler()
	do(h, http.MethodPost, "/todos", `{"title":"one"}`)
	do(h, http.MethodPost, "/todos", `{"title":"two"}`)

	rec := do(h, http.MethodGet, "/todos", "")
	todos := decodeTodos(t, rec)
	if len(todos) != 2 || todos[0].Title != "one" || todos[1].Title != "two" {
		t.Errorf("got %+v", todos)
	}
}

func TestAPI_ListCreate_TitleIsTrimmed(t *testing.T) {
	h := newHandler()
	rec := do(h, http.MethodPost, "/todos", `{"title":"  Read  "}`)
	if got := decodeTodo(t, rec).Title; got != "Read" {
		t.Errorf("title = %q, want %q", got, "Read")
	}
}

func TestAPI_ListCreate_InvalidJSON(t *testing.T) {
	h := newHandler()
	rec := do(h, http.MethodPost, "/todos", `{"title":`)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
	if msg := errorMessage(t, rec); msg != "invalid JSON" {
		t.Errorf("error = %q, want %q", msg, "invalid JSON")
	}
}

func TestAPI_ListCreate_MissingTitle(t *testing.T) {
	h := newHandler()
	for _, body := range []string{`{}`, `{"title":""}`, `{"title":"   "}`} {
		rec := do(h, http.MethodPost, "/todos", body)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("body %s: status = %d, want 400", body, rec.Code)
		}
		if msg := errorMessage(t, rec); msg != "title is required" {
			t.Errorf("body %s: error = %q, want %q", body, msg, "title is required")
		}
	}
}
```

---

## Milestone 3: Get, update and delete one todo

Implement `getTodo`, `updateTodo` and `deleteTodo`.

- A path id that is not a number gives `400` and `{"error": "invalid id"}`.
- An unknown id gives `404` and `{"error": "todo not found"}`.
- `GET /todos/{id}` returns `200` and the todo.
- `PUT /todos/{id}` takes `{"title": "...", "done": true}` and returns `200` and the updated todo. It uses the same body checks as create (`invalid JSON`, `title is required`).
- `DELETE /todos/{id}` returns `204` with an empty body.

> [!TIP]
> Write two small helpers: one that parses the id from the path, one that reads and checks the body.
> Both can write the error response themselves and return `false`. The create and update handlers can share the body helper.

**Test** `ch17/todoapi/item_test.go`

```go
package todoapi

import (
	"net/http"
	"testing"
)

func seeded() http.Handler {
	h := newHandler()
	do(h, http.MethodPost, "/todos", `{"title":"first"}`)
	return h
}

func TestAPI_Item_Get(t *testing.T) {
	h := seeded()
	rec := do(h, http.MethodGet, "/todos/1", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if got := decodeTodo(t, rec); got.ID != 1 || got.Title != "first" {
		t.Errorf("got %+v", got)
	}
}

func TestAPI_Item_GetNotFound(t *testing.T) {
	rec := do(seeded(), http.MethodGet, "/todos/99", "")
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
	if msg := errorMessage(t, rec); msg != "todo not found" {
		t.Errorf("error = %q, want %q", msg, "todo not found")
	}
}

func TestAPI_Item_InvalidID(t *testing.T) {
	for _, method := range []string{http.MethodGet, http.MethodPut, http.MethodDelete} {
		rec := do(seeded(), method, "/todos/abc", `{"title":"x"}`)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("%s: status = %d, want 400", method, rec.Code)
		}
		if msg := errorMessage(t, rec); msg != "invalid id" {
			t.Errorf("%s: error = %q, want %q", method, msg, "invalid id")
		}
	}
}

func TestAPI_Item_Update(t *testing.T) {
	h := seeded()
	rec := do(h, http.MethodPut, "/todos/1", `{"title":"changed","done":true}`)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if got := decodeTodo(t, rec); got.ID != 1 || got.Title != "changed" || !got.Done {
		t.Errorf("got %+v", got)
	}

	// the change must be visible afterwards
	again := decodeTodo(t, do(h, http.MethodGet, "/todos/1", ""))
	if again.Title != "changed" || !again.Done {
		t.Errorf("update was not saved: %+v", again)
	}
}

func TestAPI_Item_UpdateErrors(t *testing.T) {
	h := seeded()

	rec := do(h, http.MethodPut, "/todos/99", `{"title":"x"}`)
	if rec.Code != http.StatusNotFound {
		t.Errorf("unknown id: status = %d, want 404", rec.Code)
	}

	rec = do(h, http.MethodPut, "/todos/1", `{bad`)
	if rec.Code != http.StatusBadRequest || errorMessage(t, rec) != "invalid JSON" {
		t.Errorf("bad JSON: status = %d", rec.Code)
	}

	rec = do(h, http.MethodPut, "/todos/1", `{"title":" ","done":true}`)
	if rec.Code != http.StatusBadRequest || errorMessage(t, rec) != "title is required" {
		t.Errorf("blank title: status = %d", rec.Code)
	}
}

func TestAPI_Item_Delete(t *testing.T) {
	h := seeded()

	rec := do(h, http.MethodDelete, "/todos/1", "")
	if rec.Code != http.StatusNoContent {
		t.Errorf("status = %d, want 204", rec.Code)
	}
	if rec.Body.Len() != 0 {
		t.Errorf("body = %q, want empty", rec.Body.String())
	}

	if rec := do(h, http.MethodGet, "/todos/1", ""); rec.Code != http.StatusNotFound {
		t.Errorf("after delete, GET status = %d, want 404", rec.Code)
	}
	if rec := do(h, http.MethodDelete, "/todos/1", ""); rec.Code != http.StatusNotFound {
		t.Errorf("second delete status = %d, want 404", rec.Code)
	}
}
```

---

## Milestone 4: Filter by status

Extend `listTodos`.

- `GET /todos?done=true` returns only finished todos.
- `GET /todos?done=false` returns only open todos.
- Without `done`, return all.
- Any other value (for example `done=maybe`) gives `400` and `{"error": "done must be true or false"}`.
- The result is still `[]` and not `null` when nothing matches.

**Test** `ch17/todoapi/filter_test.go`

```go
package todoapi

import (
	"net/http"
	"strings"
	"testing"
)

func filterSetup() http.Handler {
	h := newHandler()
	do(h, http.MethodPost, "/todos", `{"title":"a"}`)
	do(h, http.MethodPost, "/todos", `{"title":"b"}`)
	do(h, http.MethodPost, "/todos", `{"title":"c"}`)
	do(h, http.MethodPut, "/todos/2", `{"title":"b","done":true}`)
	return h
}

func TestAPI_Filter_Done(t *testing.T) {
	rec := do(filterSetup(), http.MethodGet, "/todos?done=true", "")
	todos := decodeTodos(t, rec)
	if len(todos) != 1 || todos[0].Title != "b" {
		t.Errorf("done=true gave %+v, want only b", todos)
	}
}

func TestAPI_Filter_NotDone(t *testing.T) {
	rec := do(filterSetup(), http.MethodGet, "/todos?done=false", "")
	todos := decodeTodos(t, rec)
	if len(todos) != 2 || todos[0].Title != "a" || todos[1].Title != "c" {
		t.Errorf("done=false gave %+v, want a and c", todos)
	}
}

func TestAPI_Filter_NoFilter(t *testing.T) {
	rec := do(filterSetup(), http.MethodGet, "/todos", "")
	if todos := decodeTodos(t, rec); len(todos) != 3 {
		t.Errorf("got %d todos, want 3", len(todos))
	}
}

func TestAPI_Filter_InvalidValue(t *testing.T) {
	rec := do(filterSetup(), http.MethodGet, "/todos?done=maybe", "")
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
	if msg := errorMessage(t, rec); msg != "done must be true or false" {
		t.Errorf("error = %q", msg)
	}
}

func TestAPI_Filter_NoMatchIsEmptyArray(t *testing.T) {
	h := newHandler()
	do(h, http.MethodPost, "/todos", `{"title":"a"}`)

	rec := do(h, http.MethodGet, "/todos?done=true", "")
	if body := strings.TrimSpace(rec.Body.String()); body != "[]" {
		t.Errorf("body = %q, want []", body)
	}
}
```

<details>
<summary>Solution for milestones 2, 3 and 4 (handlers.go)</summary>

```go
package todoapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
)

type API struct {
	store *Store
}

func NewAPI(store *Store) *API {
	return &API{store: store}
}

func (a *API) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /todos", a.listTodos)
	mux.HandleFunc("POST /todos", a.createTodo)
	mux.HandleFunc("GET /todos/{id}", a.getTodo)
	mux.HandleFunc("PUT /todos/{id}", a.updateTodo)
	mux.HandleFunc("DELETE /todos/{id}", a.deleteTodo)
	return mux
}

type todoInput struct {
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

// readInput decodes and checks the body. If it returns false,
// the error response has already been written.
func readInput(w http.ResponseWriter, r *http.Request) (todoInput, bool) {
	var in todoInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return in, false
	}
	in.Title = strings.TrimSpace(in.Title)
	if in.Title == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return in, false
	}
	return in, true
}

// parseID reads {id} from the path. If it returns false,
// the error response has already been written.
func parseID(w http.ResponseWriter, r *http.Request) (int, bool) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return 0, false
	}
	return id, true
}

func (a *API) listTodos(w http.ResponseWriter, r *http.Request) {
	todos := a.store.List()

	switch r.URL.Query().Get("done") {
	case "":
		// no filter
	case "true", "false":
		want := r.URL.Query().Get("done") == "true"
		filtered := make([]Todo, 0, len(todos))
		for _, todo := range todos {
			if todo.Done == want {
				filtered = append(filtered, todo)
			}
		}
		todos = filtered
	default:
		writeError(w, http.StatusBadRequest, "done must be true or false")
		return
	}

	writeJSON(w, http.StatusOK, todos)
}

func (a *API) createTodo(w http.ResponseWriter, r *http.Request) {
	in, ok := readInput(w, r)
	if !ok {
		return
	}
	todo := a.store.Create(in.Title)
	writeJSON(w, http.StatusCreated, todo)
}

func (a *API) getTodo(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	todo, err := a.store.Get(id)
	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, todo)
}

func (a *API) updateTodo(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	in, ok := readInput(w, r)
	if !ok {
		return
	}
	todo, err := a.store.Update(id, in.Title, in.Done)
	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, todo)
}

func (a *API) deleteTodo(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	if err := a.store.Delete(id); errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
```

</details>

### Run it for real

Add `ch17/cmd/server/main.go`:

```go
package main

import (
	"log"
	"net/http"

	"golearn/ch17/todoapi"
)

func main() {
	api := todoapi.NewAPI(todoapi.NewStore())
	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", api.Routes()))
}
```

```bash
go run ./ch17/cmd/server

curl -i -X POST localhost:8080/todos -d '{"title":"Learn Go"}'
curl -i localhost:8080/todos
curl -i -X PUT localhost:8080/todos/1 -d '{"title":"Learn Go","done":true}'
curl -i -X DELETE localhost:8080/todos/1
```

## Checkpoint

- [ ] I keep the store free of HTTP code
- [ ] I check input, return `400` on bad input, and `return` after errors
- [ ] I use `404` for missing items and `204` for deletes
- [ ] I can run my API with `curl`
