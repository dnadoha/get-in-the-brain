package temp

func CToF(c float64) float64 {
	return (9 * c / 5) + 32
}

func FToC(f float64) float64 {
	return (f - 32) * 5 / 9
}
