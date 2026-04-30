package handler

import (
	"fmt"
	"strings"
	"time"

	"akasha/internal/domain"
	"github.com/gin-gonic/gin"
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

func respondError(c *gin.Context, status int, msg string) {
	c.JSON(status, gin.H{"error": msg})
}

func respondJSON(c *gin.Context, status int, data interface{}) {
	c.JSON(status, data)
}

func FormatDeps(deps []domain.Dependency) string {
	if len(deps) == 0 {
		return "ext.libraries = [\n]\n"
	}

	maxNameLen := 0
	for _, dep := range deps {
		if len(dep.Name) > maxNameLen {
			maxNameLen = len(dep.Name)
		}
	}

	var sb strings.Builder
	sb.WriteString("ext.libraries = [\n")

	for _, dep := range deps {
		padding := strings.Repeat(" ", maxNameLen-len(dep.Name)+2)
		line := fmt.Sprintf(`  "%s"%s: "%s:%s:%s",`,
			dep.Name, padding, dep.GroupID, dep.Artifact, dep.Version)

		if dep.Remark != "" {
			line += fmt.Sprintf(" // %s", dep.Remark)
		}
		line += "\n"
		sb.WriteString(line)
	}

	sb.WriteString("]\n")
	return sb.String()
}