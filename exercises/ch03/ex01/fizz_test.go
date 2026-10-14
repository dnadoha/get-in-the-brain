package fizz

import "testing"

func TestFizzBuzz(t *testing.T) {
	tests := []struct {
		in   int
		want string
	}{
		{1, "1"},
		{2, "2"},
		{3, "Fizz"},
		{5, "Buzz"},
		{7, "7"},
		{9, "Fizz"},
		{10, "Buzz"},
		{15, "FizzBuzz"},
		{30, "FizzBuzz"},
	}
	for _, tc := range tests {
		if got := FizzBuzz(tc.in); got != tc.want {
			t.Errorf("FizzBuzz(%d) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
