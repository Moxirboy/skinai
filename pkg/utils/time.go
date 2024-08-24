package utils

import "time"

func ParseDateString(dateString string) time.Time {
	layout := "2006-01-02"
	parsedDate, err := time.Parse(layout, dateString)
	if err != nil {
		return time.Time{}
	}
	return parsedDate
}

func FormatTimeToRFC3339(t time.Time) string {
	return t.Format("2006-01-02")
}
