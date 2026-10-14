package funcs

import "testing"

func TestDivide(t *testing.T) {
	got, err := Divide(10, 4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 2.5 {
		t.Errorf("Divide(10, 4) = %v, want 2.5", got)
	}
}

func TestDivideByZero(t *testing.T) {
	_, err := Divide(1, 0)
	if err == nil {
		t.Error("expected an error when dividing by zero")
	}
}

func TestMinMax(t *testing.T) {
	tests := []struct{ a, b, c, min, max int }{
		{1, 2, 3, 1, 3},
		{3, 2, 1, 1, 3},
		{2, 9, -4, -4, 9},
		{5, 5, 5, 5, 5},
	}
	for _, tc := range tests {
		min, max := MinMax(tc.a, tc.b, tc.c)
		if min != tc.min || max != tc.max {
			t.Errorf("MinMax(%d, %d, %d) = (%d, %d), want (%d, %d)",
				tc.a, tc.b, tc.c, min, max, tc.min, tc.max)
		}
	}
}
