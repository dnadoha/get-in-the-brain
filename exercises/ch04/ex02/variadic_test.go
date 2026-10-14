package variadic

import "testing"

func TestSum(t *testing.T) {
	if got := Sum(); got != 0 {
		t.Errorf("Sum() = %d, want 0", got)
	}
	if got := Sum(1, 2, 3, 4); got != 10 {
		t.Errorf("Sum(1, 2, 3, 4) = %d, want 10", got)
	}
	nums := []int{5, 5, 5}
	if got := Sum(nums...); got != 15 {
		t.Errorf("Sum(nums...) = %d, want 15", got)
	}
}

func TestMax(t *testing.T) {
	if got := Max(7); got != 7 {
		t.Errorf("Max(7) = %d, want 7", got)
	}
	if got := Max(1, 9, 3); got != 9 {
		t.Errorf("Max(1, 9, 3) = %d, want 9", got)
	}
	if got := Max(-5, -2, -9); got != -2 {
		t.Errorf("Max(-5, -2, -9) = %d, want -2", got)
	}
}
