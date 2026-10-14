package collatz

import "testing"

func TestCollatzSteps(t *testing.T) {
	tests := []struct{ in, want int }{
		{1, 0},
		{2, 1},
		{6, 8},
		{7, 16},
		{27, 111},
		{0, -1},
		{-3, -1},
	}
	for _, tc := range tests {
		if got := CollatzSteps(tc.in); got != tc.want {
			t.Errorf("CollatzSteps(%d) = %d, want %d", tc.in, got, tc.want)
		}
	}
}
