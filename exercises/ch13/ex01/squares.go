package squares

import "sync"

func ParallelSquares(nums []int) []int {
	out := make([]int, len(nums))
	var wg sync.WaitGroup

	for i, v := range nums {
		wg.Go(func() {
			out[i] = v * v
		})
	}
	wg.Wait()

	return out
}
