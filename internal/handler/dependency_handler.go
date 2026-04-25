package handler

import (
	"akasha/internal/domain"
	"akasha/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type DependencyHandler struct {
	svc        *service.DependencyService
	branchSvc  *service.BranchService
}

func NewDependencyHandler(svc *service.DependencyService, branchSvc *service.BranchService) *DependencyHandler {
	return &DependencyHandler{svc: svc, branchSvc: branchSvc}
}

func (h *DependencyHandler) List(c *gin.Context) {
	branch := c.Query("branch")
	if branch == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing branch param"})
		return
	}
	deps, err := h.svc.List(branch)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, deps)
}

func (h *DependencyHandler) Get(c *gin.Context) {
	name := c.Param("name")
	branch := c.Query("branch")
	if branch == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing branch param"})
		return
	}
	dep, err := h.svc.Get(name, branch)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, dep)
}

func (h *DependencyHandler) History(c *gin.Context) {
	name := c.Param("name")
	branch := c.Query("branch")
	if branch == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing branch param"})
		return
	}
	history, err := h.svc.History(name, branch)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, history)
}

func (h *DependencyHandler) GetAt(c *gin.Context) {
	name := c.Param("name")
	branch := c.Query("branch")
	if branch == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing branch param"})
		return
	}
	atStr := c.Query("at")
	if atStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing at param"})
		return
	}
	at, err := parseTime(atStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid at format"})
		return
	}
	dep, err := h.svc.GetAt(name, branch, at)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, dep)
}

func (h *DependencyHandler) Create(c *gin.Context) {
	var req struct {
		Name     string `json:"name" binding:"required"`
		GroupID  string `json:"groupId" binding:"required"`
		Artifact string `json:"artifact" binding:"required"`
		Version  string `json:"version" binding:"required"`
		Branch   string `json:"branch" binding:"required"`
		SourceIP string `json:"sourceIp"`
		Remark   string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// Check if branch can be modified
	if !h.branchSvc.CanModify(req.Branch) {
		c.JSON(http.StatusForbidden, gin.H{"error": "branch is archived or deleted, cannot modify"})
		return
	}
	d := &domain.Dependency{
		Name:     req.Name,
		GroupID:  req.GroupID,
		Artifact: req.Artifact,
		Version:  req.Version,
		Branch:   req.Branch,
		SourceIP: req.SourceIP,
		Remark:   req.Remark,
	}
	if err := h.svc.CreateOrUpdate(d); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, d)
}

func (h *DependencyHandler) Delete(c *gin.Context) {
	name := c.Param("name")
	branch := c.Query("branch")
	if branch == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing branch param"})
		return
	}
	if err := h.svc.Delete(name, branch); err != nil {
		if _, ok := err.(*service.ErrForbidden); ok {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// GetDepsAt 批量闪回查询 - 查询某分支在某时间点的所有依赖
func (h *DependencyHandler) GetDepsAt(c *gin.Context) {
	name := c.Param("name")
	atStr := c.Query("at")
	if atStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing at param"})
		return
	}
	at, err := parseTime(atStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid at format"})
		return
	}
	deps, err := h.svc.GetDepsAt(name, at)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, deps)
}

func (h *DependencyHandler) Compare(c *gin.Context) {
	sourceBranch := c.Query("source")
	targetBranch := c.Query("target")
	if sourceBranch == "" || targetBranch == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing source or target param"})
		return
	}

	sourceDeps, err := h.svc.List(sourceBranch)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	targetDeps, err := h.svc.List(targetBranch)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	sourceMap := make(map[string]domain.Dependency)
	for _, dep := range sourceDeps {
		sourceMap[dep.Name] = dep
	}
	targetMap := make(map[string]domain.Dependency)
	for _, dep := range targetDeps {
		targetMap[dep.Name] = dep
	}

	type DiffItem struct {
		Name        string `json:"name"`
		SourceCoord string `json:"sourceCoord"`
		TargetCoord string `json:"targetCoord"`
		Type        string `json:"type"`
	}

	var diffs []DiffItem
	var added, removed, modified, unchanged int

	for name, sourceDep := range sourceMap {
		if targetDep, exists := targetMap[name]; exists {
			if sourceDep.Version == targetDep.Version {
				diffs = append(diffs, DiffItem{
					Name:        name,
					SourceCoord: sourceDep.MavenCoord(),
					TargetCoord: targetDep.MavenCoord(),
					Type:        "unchanged",
				})
				unchanged++
			} else {
				diffs = append(diffs, DiffItem{
					Name:        name,
					SourceCoord: sourceDep.MavenCoord(),
					TargetCoord: targetDep.MavenCoord(),
					Type:        "modified",
				})
				modified++
			}
		} else {
			diffs = append(diffs, DiffItem{
				Name:        name,
				SourceCoord: sourceDep.MavenCoord(),
				TargetCoord: "",
				Type:        "removed",
			})
			removed++
		}
	}

	for name, targetDep := range targetMap {
		if _, exists := sourceMap[name]; !exists {
			diffs = append(diffs, DiffItem{
				Name:        name,
				SourceCoord: "",
				TargetCoord: targetDep.MavenCoord(),
				Type:        "added",
			})
			added++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"sourceBranch": sourceBranch,
		"targetBranch": targetBranch,
		"diffs":        diffs,
		"summary": gin.H{
			"added":     added,
			"removed":   removed,
			"modified":  modified,
			"unchanged": unchanged,
			"total":     len(diffs),
		},
	})
}

// HistoryBetween 获取某时间段内的依赖变更
func (h *DependencyHandler) HistoryBetween(c *gin.Context) {
	name := c.Param("name")
	branch := c.Query("branch")
	if branch == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing branch param"})
		return
	}
	startAtStr := c.Query("startAt")
	endAtStr := c.Query("endAt")
	if startAtStr == "" || endAtStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing startAt or endAt param"})
		return
	}
	startAt, err := parseTime(startAtStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid startAt format"})
		return
	}
	endAt, err := parseTime(endAtStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid endAt format"})
		return
	}
	deps, err := h.svc.HistoryBetween(name, branch, startAt, endAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, deps)
}