package date

import (
	"time"

	"github.com/svandecappelle/gitcontrib/internal/interfaces"
)

// TODO get a better idea
const OutOfRange = 99999

// GetBeginningOfDay given a time.Time calculates the start time of that day
func GetBeginningOfDay(t time.Time) time.Time {
	year, month, day := t.Date()
	startOfDay := time.Date(year, month, day, 0, 0, 0, 0, t.Location())
	return startOfDay
}

// GetEndOfDay given a time.Time calculates the end time of that day
func GetEndOfDay(t time.Time) time.Time {
	year, month, day := t.Date()
	startOfDay := time.Date(year, month, day, 23, 59, 59, 0, t.Location())
	return startOfDay
}

// countDaysSinceDate counts how many days passed since the passed `date`
func CountDaysSinceDate(date time.Time, r *interfaces.StatsResult) int {
	days := 0
	endDate := GetEndOfDay(r.EndOfScan)
	if !date.Before(endDate) && !date.Equal(endDate) {
		return OutOfRange
	} else if date.Equal(endDate) {
		return days
	}
	for date.Before(endDate) {
		date = date.Add(time.Hour * 24)
		days++
		if days > r.DurationInDays {
			return OutOfRange
		}
	}
	return days
}

func DaysBetween(begin time.Time, end time.Time) int {
	return int(end.Sub(begin).Hours() / 24)
}
