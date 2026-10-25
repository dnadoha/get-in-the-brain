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
