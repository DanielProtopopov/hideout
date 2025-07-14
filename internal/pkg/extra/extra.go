package extra

import (
	"time"
)

func TruncateToDay(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
}

func TruncateToEndDay(date time.Time) time.Time {
	dateBeginDay := TruncateToDay(date)
	return time.Date(dateBeginDay.Year(), dateBeginDay.Month(), dateBeginDay.Day(), 23, 59, 59, 999, time.UTC)
}

func TruncateToMonth(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, time.UTC)
}

func TruncateToEndMonth(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), daysIn(date.Month(), date.Year()), 23, 59, 59, 999, time.UTC)
}

func daysIn(m time.Month, year int) int {
	return time.Date(year, m+1, 0, 0, 0, 0, 0, time.UTC).Day()
}
