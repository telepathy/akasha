package handler

import (
	"net/http"

	"akasha/internal/domain"
	"akasha/internal/service"

	"github.com/gin-gonic/gin"
)

type GradleHandler struct {
	svc       *service.DependencyService
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

	var result = "ext.libraries = [\n"
	result += `/**` + "\n*下面是二方包\n**/\n"

	for _, dep := range deps {
		result += `"` + dep.Name + `"` + "                                                        : "
		result += `"` + dep.GroupID + ":" + dep.Artifact + ":" + dep.Version + `",`
		result += " // " + dep.CreatedAt.Format("2006-01-02T15:04:05")
		if dep.SourceIP != "" {
			result += ", " + dep.SourceIP
		}
		if dep.Remark != "" {
			result += ",备注： " + dep.Remark
		}
		result += "\n"
	}
	result += "]\n"
	return result
}