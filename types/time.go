package types

import (
	"strconv"
	"time"
)

// TimeBod returns beginning of the day given.
func TimeBod(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}

// TimeBom returns beginning of the month given.
func TimeBom(t time.Time) time.Time {
	year, month, _ := t.Date()
	return time.Date(year, month, 1, 0, 0, 0, 0, t.Location())
}

// TimeMonSunWeekDay converts week day to Monday(0) - Sunday(6) index
func TimeMonSunWeekDay(t time.Time) int {
	current := int(t.Weekday())
	// Convert weekday value to Monday-Sunday
	current = current - 1
	if current == -1 {
		current = 6
	}
	return current
}

// DateEqual takes two time and checks date (year, month, day) equality of them.
func DateEqual(date1, date2 time.Time) bool {
	y1, m1, d1 := date1.Date()
	y2, m2, d2 := date2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

//  ParseUnix parses the Millisecond given and returns unix Time. (momentjs gives ms as string)
func ParseUnix(msString string) (time.Time, error) {
	i, err := strconv.ParseInt(msString, 10, 64)
	return time.Unix(0, i*1000000), err
}

// DaysInMonth returns the count of days in a month
func DaysInMonth(m time.Month, year int) int {
	return time.Date(year, m+1, 0, 0, 0, 0, 0, time.UTC).Day()
}
