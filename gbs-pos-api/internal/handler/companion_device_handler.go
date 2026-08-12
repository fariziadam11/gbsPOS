package handler

import (
	"net/http"
	"strings"
	"time"

	"gbs-pos-api/internal/model"
	"gbs-pos-api/internal/repository"

	"github.com/gin-gonic/gin"
)

type CompanionDeviceHandler struct {
	repo *repository.CompanionDeviceRepository
}

func NewCompanionDeviceHandler(repo *repository.CompanionDeviceRepository) *CompanionDeviceHandler {
	return &CompanionDeviceHandler{repo: repo}
}

type registerCompanionRequest struct {
	DeviceID     string   `json:"deviceId" binding:"required"`
	DeviceName   string   `json:"deviceName"`
	SDKVersion   string   `json:"sdkVersion"`
	Capabilities []string `json:"capabilities"`
}

func (h *CompanionDeviceHandler) Register(c *gin.Context) {
	var req registerCompanionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "VALIDATION_ERROR", "message": err.Error()})
		return
	}
	device := &model.CompanionDevice{
		DeviceID:     req.DeviceID,
		DeviceName:   req.DeviceName,
		SDKVersion:   req.SDKVersion,
		Capabilities: strings.Join(req.Capabilities, ","),
		LastSeenAt:   time.Now(),
	}
	if err := h.repo.Upsert(c.Request.Context(), device); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "DEVICE_REGISTER_ERROR"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": device})
}
