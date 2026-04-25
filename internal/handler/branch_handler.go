package handler

import (
	"akasha/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type BranchHandler struct {
	svc     *service.BranchService
	depSvc  *service.DependencyService
}

func NewBranchHandler(svc *service.BranchService, depSvc *service.DependencyService) *BranchHandler {
	return &BranchHandler{svc: svc, depSvc: depSvc}
}

func (h *BranchHandler) List(c *gin.Context) {
	branches, err := h.svc.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, branches)
}

func (h *BranchHandler) Get(c *gin.Context) {
	name := c.Param("name")
	branch, err := h.svc.Get(name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, branch)
}

func (h *BranchHandler) Create(c *gin.Context) {
	var req struct {
		Name       string `json:"name" binding:"required"`
		BaseBranch string `json:"baseBranch"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.Create(req.Name, req.BaseBranch); err != nil {
		if _, ok := err.(*service.ErrBranchExists); ok {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "created"})
}

func (h *BranchHandler) Delete(c *gin.Context) {
	name := c.Param("name")
	if err := h.svc.Delete(name); err != nil {
		if _, ok := err.(*service.ErrForbidden); ok {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (h *BranchHandler) Merge(c *gin.Context) {
	name := c.Param("name")
	var req struct {
		TargetBranch string `json:"targetBranch" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.Merge(name, req.TargetBranch); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "merged"})
}

func (h *BranchHandler) Archive(c *gin.Context) {
	name := c.Param("name")
	if err := h.svc.Archive(name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "archived"})
}

func (h *BranchHandler) Unlock(c *gin.Context) {
	name := c.Param("name")
	if err := h.svc.Unlock(name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "unlocked"})
}

func (h *BranchHandler) GetHistory(c *gin.Context) {
	name := c.Param("name")
	history, err := h.svc.History(name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, history)
}

func (h *BranchHandler) GetDepsText(c *gin.Context) {
	name := c.Param("name")
	deps, err := h.depSvc.List(name)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	output := formatDeps(deps)
	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.String(http.StatusOK, output)
}