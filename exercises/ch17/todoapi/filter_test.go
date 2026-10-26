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
