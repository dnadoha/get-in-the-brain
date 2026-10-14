package variadic

func Sum(nums ...int) int {
	sum := 0
	for _, n := range nums {
		sum += n
	}
	return sum
}

func Max(first int, rest ...int) int {
	mx := first
	for _, n := range rest {
		if n > mx {
			mx = n
		}
	}

	return mx
}
