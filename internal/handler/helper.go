package handler

import (
	"time"

	"akasha/internal/domain"
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

func FormatDeps(deps []domain.Dependency) string {
	if len(deps) == 0 {
		return "ext.libraries = [\n]\n"
	}
	result := "ext.libraries = [\n"
	for _, dep := range deps {
		result += `"` + dep.Name + `": "` + dep.GroupID + ":" + dep.Artifact + ":" + dep.Version + `",`
		if dep.Remark != "" {
			result += " // " + dep.Remark
		}
		result += "\n"
	}
	result += "]\n"
	return result
}