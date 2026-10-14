package grade

import "testing"

func TestGrade(t *testing.T) {
	tests := []struct {
		score int
		want  string
	}{
		{100, "A"}, {90, "A"},
		{89, "B"}, {80, "B"},
		{79, "C"}, {70, "C"},
		{69, "D"}, {60, "D"},
		{59, "F"}, {0, "F"},
		{-1, "Invalid"}, {101, "Invalid"},
	}
	for _, tc := range tests {
		if got := Grade(tc.score); got != tc.want {
			t.Errorf("Grade(%d) = %q, want %q", tc.score, got, tc.want)
		}
	}
}
