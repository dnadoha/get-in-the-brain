package collatz

func CollatzSteps(n int) int {
	if n <= 0 {
		return -1
	}

	step_counter := 0
	for n != 1 {
		if n%2 == 0 {
			n /= 2
		} else {
			n = 3*n + 1
		}
		step_counter++
	}

	return step_counter
}
