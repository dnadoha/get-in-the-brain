package grade

func Grade(score int) string {
	switch {
	case score < 0 || score > 100:
		return "Invalid"
	case score >= 90:
		return "A"
	case score >= 80:
		return "B"
	case score >= 70:
		return "C"
	case score >= 60:
		return "D"
	}
	return "F"
}
