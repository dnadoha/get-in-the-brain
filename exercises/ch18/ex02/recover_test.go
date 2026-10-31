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
