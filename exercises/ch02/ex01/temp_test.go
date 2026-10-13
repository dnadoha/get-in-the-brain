package temp

import (
	"math"
	"testing"
)

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) < 0.0001
}

func TestCToF(t *testing.T) {
	tests := []struct{ c, f float64 }{
		{0, 32},
		{100, 212},
		{-40, -40},
	}

	for _, tc := range tests {
		if got := CToF(tc.c); !almostEqual(got, tc.f) {
			t.Errorf("got = %v, want %v", got, tc.f)
		}
	}
}

func TestFToC(t *testing.T) {
	tests := []struct{ f, c float64 }{
		{212, 100},
		{32, 0},
	}

	for _, tc := range tests {
		if got := FToC(tc.f); got != tc.c {
			t.Errorf("got = %v, want = %v", got, tc.c)
		}
	}
}
