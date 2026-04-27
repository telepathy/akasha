package handler

import (
	"akasha/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type InitHandler struct {
	svc *service.InitService
}

func NewInitHandler(svc *service.InitService) *InitHandler {
	return &InitHandler{svc: svc}
}

func (h *InitHandler) HealthDB(c *gin.Context) {
	status, err := h.svc.CheckDBStatus()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, status)
}

func (h *InitHandler) InitDB(c *gin.Context) {
	status, err := h.svc.Initialize()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, status)
}