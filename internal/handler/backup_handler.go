package handler

import (
	"akasha/internal/service"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type BackupHandler struct {
	svc *service.BackupService
}

func NewBackupHandler(svc *service.BackupService) *BackupHandler {
	return &BackupHandler{svc: svc}
}

func (h *BackupHandler) Export(c *gin.Context) {
	data, err := h.svc.Export()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	filename := "akasha-backup-" + time.Now().Format("20060102-150405") + ".json"
	c.Header("Content-Type", "application/json")
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.JSON(http.StatusOK, data)
}

func (h *BackupHandler) Restore(c *gin.Context) {
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing file field"})
		return
	}
	defer file.Close()

	raw, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read file"})
		return
	}

	data, err := h.svc.ValidateJSON(raw)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.Restore(data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"restored":     true,
		"branches":     len(data.Branches),
		"dependencies": len(data.Dependencies),
	})
}
