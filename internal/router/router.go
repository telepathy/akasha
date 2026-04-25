package router

import (
	"net/http"

	"akasha/internal/domain"
	"akasha/internal/handler"
	"akasha/internal/service"

	"github.com/gin-gonic/gin"
)

func Setup(depSvc *service.DependencyService, branchSvc *service.BranchService, gradlePassword string) *gin.Engine {
	r := gin.Default()

	depHandler := handler.NewDependencyHandler(depSvc, branchSvc)
	branchHandler := handler.NewBranchHandler(branchSvc, depSvc)
	_ = gradlePassword

	r.LoadHTMLGlob("templates/*")
	r.Static("/static", "static")

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	r.GET("/branches", func(c *gin.Context) {
		c.HTML(http.StatusOK, "branches.html", nil)
	})

	r.GET("/compare", func(c *gin.Context) {
		c.HTML(http.StatusOK, "compare.html", nil)
	})

	r.GET("/merge", func(c *gin.Context) {
		c.HTML(http.StatusOK, "merge.html", nil)
	})

	r.GET("/dependencies", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	r.GET("/dependency", func(c *gin.Context) {
		branchName := c.Query("branch")
		if branchName == "" {
			branchName = "main"
		}
		// Check if branch is deleted
		if branchSvc.IsDeleted(branchName) {
			c.String(http.StatusForbidden, "branch is deleted")
			return
		}
		deps, err := depSvc.List(branchName)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		output := formatDeps(deps)
		c.Header("Content-Type", "text/plain; charset=utf-8")
		c.Header("Content-Disposition", "attachment; filename=dependency.gradle")
		c.String(http.StatusOK, output)
	})

	api := r.Group("/api/v1")
	{
		api.GET("/dependencies", depHandler.List)
		api.GET("/dependencies/:name", depHandler.Get)
		api.GET("/dependencies/:name/history", depHandler.History)
		api.GET("/dependencies/:name/at", depHandler.GetAt)
		api.GET("/dependencies/:name/history-between", depHandler.HistoryBetween)
		api.POST("/dependencies", depHandler.Create)
		api.DELETE("/dependencies/:name", depHandler.Delete)
		api.GET("/dependencies/compare", depHandler.Compare)

		// 分支批量操作
		api.GET("/branches", branchHandler.List)
		api.GET("/branches/:name", branchHandler.Get)
		api.POST("/branches", branchHandler.Create)
		api.DELETE("/branches/:name", branchHandler.Delete)
		api.POST("/branches/:name/merge", branchHandler.Merge)
		api.POST("/branches/:name/archive", branchHandler.Archive)
		api.POST("/branches/:name/unlock", branchHandler.Unlock)

		// 批量闪回和历史查询
		api.GET("/branches/:name/deps-at", depHandler.GetDepsAt)
		api.GET("/branches/:name/history", branchHandler.GetHistory)
		api.GET("/branches/:name/deps-text", branchHandler.GetDepsText)
	}

	return r
}

func formatDeps(deps []domain.Dependency) string {
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