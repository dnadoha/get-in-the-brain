package greet

import (
	"testing"
)

func TestGreet(t *testing.T) {
	want := "Hello, Ana!"
	if got := Greet("Ana"); got != want {
		t.Errorf("got = %q, want = %q", got, want)
	}
}

func TestGreetEmpty(t *testing.T) {
	want := "Hello, World!"
	if got := Greet(""); got != want {
		t.Errorf("got = %q, want = %q", got, want)
	}
}
