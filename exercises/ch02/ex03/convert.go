package convert

import "math"

func Percent(part, total int) float64 {
	if total == 0 {
		return float64(0)
	}
	return float64(part) * 100 / float64(total)
}

func ToInt8(n int) (int8, bool) {
	if n > math.MaxInt8 || n < math.MinInt8 {
		return 0, false
	}
	return int8(n), true
}
