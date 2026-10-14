package funcs

import (
	"errors"
)

func Divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, errors.New("cannot divide by 0")
	}

	return a / b, nil
}

func MinMax(a, b, c int) (min, max int) {
	mn := a
	mx := c

	if b < a {
		mn = b
	}
	if c < b {
		mn = c
	}

	if b > c {
		mx = b
	}
	if a > b {
		mx = a
	}

	return mn, mx
}
