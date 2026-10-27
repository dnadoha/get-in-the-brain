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
