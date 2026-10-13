package double

import (
	"testing"
)

func TestDouble(t *testing.T) {
	want1 := 4
	if got := Double(2); got != want1 {
		t.Errorf("got = %d, want = %d", got, want1)
	}

	want2 := -44
	if got := Double(-22); got != want2 {
		t.Errorf("got = %d, want = %d", got, want2)
	}

	want3 := 0
	if got := Double(0); got != want3 {
		t.Errorf("got = %d, want = %d", got, want3)
	}
}
