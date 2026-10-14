package closure

import "testing"

func TestCounter(t *testing.T) {
	c := Counter()
	for want := 1; want <= 3; want++ {
		if got := c(); got != want {
			t.Errorf("call %d returned %d", want, got)
		}
	}
}

func TestCountersAreIndependent(t *testing.T) {
	a := Counter()
	b := Counter()
	a()
	a()
	if got := b(); got != 1 {
		t.Errorf("b() = %d, want 1 (counters must not share state)", got)
	}
}

func TestFibonacci(t *testing.T) {
	f := Fibonacci()
	want := []int{0, 1, 1, 2, 3, 5, 8, 13, 21, 34}
	for i, w := range want {
		if got := f(); got != w {
			t.Errorf("call %d = %d, want %d", i+1, got, w)
		}
	}
}
