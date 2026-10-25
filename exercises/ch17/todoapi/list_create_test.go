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
