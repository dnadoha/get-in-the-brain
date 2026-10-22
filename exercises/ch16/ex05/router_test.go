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
