package handler

import (
	"time"

	"gorm.io/gorm"
)

func parseTime(s string) (time.Time, error) {
	for _, fmt := range []string{
		time.RFC3339,
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
		"2006-01-02",
	} {
		if t, err := time.Parse(fmt, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, gorm.ErrInvalidValue
}