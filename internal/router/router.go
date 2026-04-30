package router

import (
	"embed"
	"html/template"
	"io/fs"
	"net/http"

	"akasha/internal/handler"
	"akasha/internal/service"

	"github.com/gin-gonic/gin"
)

type RouterConfig struct {
	Password     string
	APIKey       string
	JWTSecret    string
	ExternalHost string
	TemplateFS   embed.FS
	StaticFS     embed.FS
}

func Setup(depSvc *service.DependencyService, branchSvc *service.BranchService, initSvc *service.InitService, cfg RouterConfig) *gin.Engine {
	r := gin.Default()

	depHandler := handler.NewDependencyHandler(depSvc, branchSvc)
	branchHandler := handler.NewBranchHandler(branchSvc, depSvc)
	initHandler := handler.NewInitHandler(initSvc)
	authHandler := handler.NewAuthHandler(cfg.Password, cfg.APIKey, cfg.JWTSecret)

	templ := template.Must(template.New("").ParseFS(cfg.TemplateFS, "templates/*.html"))
	r.SetHTMLTemplate(templ)
	staticSub, _ := fs.Sub(cfg.StaticFS, "static")
	r.StaticFS("/static", http.FS(staticSub))

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{"host": cfg.ExternalHost})
	})

	r.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", nil)
	})

	r.GET("/dependencies", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	r.GET("/dependency", func(c *gin.Context) {
		branchName := c.Query("branch")
		if branchName == "" {
			branchName = "main"
		}
		if branchSvc.IsDeleted(branchName) {
			c.String(http.StatusForbidden, "branch is deleted")
			return
		}
		deps, err := depSvc.List(branchName)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		output := handler.FormatDeps(deps)
		c.Header("Content-Type", "text/plain; charset=utf-8")
		c.Header("Content-Disposition", "attachment; filename=dependency.gradle")
		c.String(http.StatusOK, output)
	})

	r.GET("/compare", func(c *gin.Context) {
		c.HTML(http.StatusOK, "compare.html", nil)
	})

	protected := r.Group("/", authHandler.Middleware())
	{
		protected.GET("/branches", func(c *gin.Context) {
			c.HTML(http.StatusOK, "branches.html", nil)
		})
		protected.GET("/merge", func(c *gin.Context) {
			c.HTML(http.StatusOK, "merge.html", nil)
		})
	}

	api := r.Group("/api/v1")
	{
		api.GET("/dependencies", depHandler.List)
		api.GET("/dependencies/:name", depHandler.Get)
		api.GET("/dependencies/:name/history", depHandler.History)
		api.GET("/dependencies/:name/at", depHandler.GetAt)
		api.GET("/dependencies/:name/history-between", depHandler.HistoryBetween)
		api.GET("/dependencies/compare", depHandler.Compare)

		api.GET("/branches", branchHandler.List)
		api.GET("/branches/:name", branchHandler.Get)
		api.GET("/branches/:name/deps-at", depHandler.GetDepsAt)
		api.GET("/branches/:name/history", branchHandler.GetHistory)
		api.GET("/branches/:name/deps-text", branchHandler.GetDepsText)

		api.GET("/health/db", initHandler.HealthDB)

		api.POST("/login", authHandler.Login)
		api.POST("/logout", authHandler.Logout)

		write := api.Group("/", authHandler.RequireAuth())
		{
			write.POST("/dependencies", depHandler.Create)
			write.DELETE("/dependencies/:name", depHandler.Delete)

			write.POST("/branches", branchHandler.Create)
			write.DELETE("/branches/:name", branchHandler.Delete)
			write.POST("/branches/:name/merge", branchHandler.Merge)
			write.POST("/branches/:name/archive", branchHandler.Archive)
			write.POST("/branches/:name/unlock", branchHandler.Unlock)

			write.POST("/init", initHandler.InitDB)
		}
	}

	return r
}