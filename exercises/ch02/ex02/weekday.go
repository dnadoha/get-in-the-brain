package weekday

type Weekday int

// TODO: add the constants Sunday ... Saturday using iota
const (
	Sunday Weekday = iota
	Monday
	Tuesday
	Wednesday
	Thursday
	Friday
	Saturday
)

var days = [...]string{
	"Sunday", "Monday", "Tuesday",
	"Wednesday", "Thursday", "Friday",
	"Saturday",
}

func (d Weekday) String() string {
	if d < Sunday || d > Saturday {
		return "Unknown"
	}
	return days[d]
}

func (d Weekday) IsWeekend() bool {
	return d == Saturday || d == Sunday
}
