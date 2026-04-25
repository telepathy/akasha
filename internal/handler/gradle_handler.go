package handler

import (
	"fmt"
	"net/http"
	"strings"

	"akasha/internal/domain"
	"akasha/internal/service"

	"github.com/gin-gonic/gin"
)

type GradleHandler struct {
	svc      *service.DependencyService
	password string
}

func NewGradleHandler(svc *service.DependencyService, password string) *GradleHandler {
	return &GradleHandler{svc: svc, password: password}
}

func (h *GradleHandler) Output(c *gin.Context) {
	branch := c.Param("branch")
	if branch == "" {
		c.String(http.StatusBadRequest, "missing branch")
		return
	}

	pwd := c.Query("pwd")
	if pwd != h.password {
		c.String(http.StatusUnauthorized, "invalid password")
		return
	}

	deps, err := h.svc.List(branch)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	output := formatDeps(deps)
	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.String(http.StatusOK, output)
}

func formatDeps(deps []domain.Dependency) string {
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
		line := fmt.Sprintf(`  "%s"%s: "%s:%s:%s"`,
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