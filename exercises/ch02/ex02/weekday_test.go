package weekday

import (
	"fmt"
	"testing"
)

func TestValues(t *testing.T) {
	if Sunday != 0 || Monday != 1 || Saturday != 6 {
		t.Fatalf("wrong values: Sunday=%d Monday=%d Saturday=%d", Sunday, Monday, Saturday)
	}
}

func TestString(t *testing.T) {
	tests := map[Weekday]string{
		Sunday:     "Sunday",
		Monday:     "Monday",
		Wednesday:  "Wednesday",
		Saturday:   "Saturday",
		Weekday(9): "Unknown",
	}
	for day, want := range tests {
		if got := day.String(); got != want {
			t.Errorf("%d.String() = %q, want %q", int(day), got, want)
		}
	}
}

func TestPrintUsesString(t *testing.T) {
	if got := fmt.Sprint(Friday); got != "Friday" {
		t.Errorf("Sprint(Friday) = %q, want Friday", got)
	}
}

func TestIsWeekend(t *testing.T) {
	if !Saturday.IsWeekend() || !Sunday.IsWeekend() {
		t.Error("Saturday and Sunday should be weekend days")
	}
	if Monday.IsWeekend() || Friday.IsWeekend() {
		t.Error("Monday and Friday should not be weekend days")
	}
}
