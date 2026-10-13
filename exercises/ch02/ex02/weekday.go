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

func (d Weekday) String() string {
	return "" // TODO
}

func (d Weekday) IsWeekend() bool {
	return false // TODO
}
