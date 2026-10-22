
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
