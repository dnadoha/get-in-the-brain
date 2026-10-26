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
