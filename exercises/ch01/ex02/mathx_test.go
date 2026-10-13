package mathx

import "testing"

func TestAdd(t *testing.T) {
	want1 := 5
	if got := Add(2, 3); got != want1 {
		t.Errorf("got = %d, want = %d", got, want1)
	}

	want2 := -3
	if got := Add(-4, 1); got != want2 {
		t.Errorf("got = %d, want = %d", got, want2)
	}
}

func TestAbs(t *testing.T) {
	want1 := 7
	want2 := 7
	want3 := 0

	if got := Abs(-7); got != want1 {
		t.Errorf("got = %d, want = %d", got, want1)
	}
	if got := Abs(7); got != want2 {
		t.Errorf("got = %d, want = %d", got, want2)
	}
	if got := Abs(0); got != want3 {
		t.Errorf("got = %d, want = %d", got, want3)
	}

}
