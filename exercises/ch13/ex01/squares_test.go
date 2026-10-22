package squares

import (
	"reflect"
	"testing"
)

func TestParallelSquares(t *testing.T) {
	got := ParallelSquares([]int{1, 2, 3, 4})
	want := []int{1, 4, 9, 16}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestParallelSquaresEmpty(t *testing.T) {
	if got := ParallelSquares(nil); len(got) != 0 {
		t.Errorf("got %v, want empty", got)
	}
}

func TestParallelSquaresMany(t *testing.T) {
	in := make([]int, 1000)
	want := make([]int, 1000)
	for i := range in {
		in[i] = i
		want[i] = i * i
	}
	if got := ParallelSquares(in); !reflect.DeepEqual(got, want) {
		t.Error("result is wrong or in the wrong order")
	}
}
