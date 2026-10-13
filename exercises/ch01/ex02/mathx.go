package mathx

func Add(a, b int) int {
	return a + b
}

func Abs(n int) int {
	if n < 0 {
		return n * -1
	}
	return n
}
