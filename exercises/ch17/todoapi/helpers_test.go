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
