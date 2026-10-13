package convert

import (
	"math"
	"testing"
)

func TestPercent(t *testing.T) {
	if got := Percent(1, 4); math.Abs(got-25) > 0.0001 {
		t.Errorf("Percent(1, 4) = %v, want 25", got)
	}
	if got := Percent(1, 3); math.Abs(got-33.3333) > 0.001 {
		t.Errorf("Percent(1, 3) = %v, want about 33.33", got)
	}
	if got := Percent(5, 0); got != 0 {
		t.Errorf("Percent(5, 0) = %v, want 0", got)
	}
}

func TestToInt8(t *testing.T) {
	tests := []struct {
		in   int
		want int8
		ok   bool
	}{
		{0, 0, true},
		{127, 127, true},
		{-128, -128, true},
		{128, 0, false},
		{-129, 0, false},
	}
	for _, tc := range tests {
		got, ok := ToInt8(tc.in)
		if got != tc.want || ok != tc.ok {
			t.Errorf("ToInt8(%d) = (%d, %v), want (%d, %v)", tc.in, got, ok, tc.want, tc.ok)
		}
	}
}
