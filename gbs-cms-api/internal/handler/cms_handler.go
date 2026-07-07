package handler

import (
	"errors"
	"fmt"
	"gbs-cms-api/internal/model"
	"gbs-cms-api/internal/service"
	"gbs-common/pkg/response"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CMSHandler struct {
	cmsService  *service.CMSService
	userService *service.UserService
}

func NewCMSHandler(cmsService *service.CMSService, userService *service.UserService) *CMSHandler {
	return &CMSHandler{cmsService: cmsService, userService: userService}
}

func (h *CMSHandler) resolveCreatedBy(c *gin.Context) (uint, error) {
	userID, ok := c.Get("userID")
	if !ok {
		return 0, fmt.Errorf("missing userID claim")
	}

	// Local JWT tokens set sub as a numeric user ID.
	if numericID, ok := userID.(float64); ok {
		return uint(numericID), nil
	}

	// Keycloak tokens set sub as a string UUID; resolve via username.
	username := c.GetString("username")
	if username == "" {
		return 0, fmt.Errorf("missing username claim")
	}
	user, err := h.userService.FindByUsername(username)
	if err != nil {
		return 0, fmt.Errorf("user not found")
	}
	return user.ID, nil
}

// UploadAd godoc
//
//	@Summary		Upload advertisement
//	@Description	Upload a new video advertisement. Only ADMIN role can upload ads.
//	@Tags			Ads
//	@Accept		multipart/form-data
//	@Produce		json
//	@Security		BearerAuth
//	@Param			file		formData	file	true	"Video file (mp4, webm, quicktime)"
//	@Param			name		formData	string	true	"Ad name"
//	@Param			storeTypes	formData	[]string	false	"Store types"	collectionFormat(multi)
//	@Param			playlistOrder	formData	int		false	"Playlist order"
//	@Param			startDate	formData	string	false	"Start date (YYYY-MM-DD)"
//	@Param			endDate		formData	string	false	"End date (YYYY-MM-DD)"
//	@Success		201
//	@Failure		400
//	@Failure		401
//	@Failure		403
//	@Failure		413
//	@Router			/v1/ads/upload [post]
func (h *CMSHandler) UploadAd(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("VALIDATION_ERROR", "File is required"))
		return
	}
	defer file.Close()
	if err := h.cmsService.ValidateUploadFile(header.Filename, header.Size); err != nil {
		switch err.Error() {
		case "FILE_TOO_LARGE":
			c.JSON(
				http.StatusRequestEntityTooLarge,
				response.Error("FILE_TOO_LARGE", "Maximum file size is 50MB"),
			)
		case "INVALID_FILE_TYPE":
			c.JSON(
				http.StatusUnsupportedMediaType,
				response.Error(
					"INVALID_FILE_TYPE",
					"Only video/mp4, video/webm, video/quicktime files are allowed",
				),
			)
		default:
			c.JSON(http.StatusBadRequest, response.Error("VALIDATION_ERROR", err.Error()))
		}
		return
	}
	name := c.PostForm("name")
	if name == "" {
		c.JSON(http.StatusUnprocessableEntity, response.ValidationError("Name is required", nil))
		return
	}
	storeTypes := c.PostFormArray("storeTypes")
	if len(storeTypes) == 0 {
		storeTypes = []string{"RETAIL"}
	}
	playlistOrder, err := service.ParseInt(c.PostForm("playlistOrder"))
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, response.ValidationError(err.Error(), nil))
		return
	}
	startDate := service.ParseDatePointer(c.PostForm("startDate"))
	endDate := service.ParseDatePointer(c.PostForm("endDate"))
	startTime := service.ParseTimePointer(c.PostForm("startTime"))
	endTime := service.ParseTimePointer(c.PostForm("endTime"))
	createdBy, err := h.resolveCreatedBy(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.Error("UNAUTHORIZED", err.Error()))
		return
	}
	ad, err := h.cmsService.CreateAd(
		name,
		header.Filename,
		header.Header.Get("Content-Type"),
		header.Size,
		storeTypes,
		playlistOrder,
		startDate,
		endDate,
		startTime,
		endTime,
		createdBy,
	)
	if err != nil {
		if err.Error() == "INVALID_SCHEDULE" {
			c.JSON(
				http.StatusUnprocessableEntity,
				response.Error(
					"INVALID_SCHEDULE",
					"Start date must be before or equal to end date",
				),
			)
			return
		}
		c.JSON(http.StatusInternalServerError, response.Error("UPLOAD_FAILED", err.Error()))
		return
	}
	if err := h.cmsService.SaveUpload(file, ad.StoragePath); err != nil {
		_ = h.cmsService.DeleteAd(ad.ID)
		c.JSON(http.StatusInternalServerError, response.Error("UPLOAD_FAILED", err.Error()))
		return
	}
	c.JSON(http.StatusCreated, response.Success(ad))
}

// ListAds godoc
//
//	@Summary		List all advertisements
//	@Description	Get paginated list of all advertisements. Only ADMIN role.
//	@Tags			Ads
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page	query		int	false	"Page number (default: 1)"
//	@Param			limit	query		int	false	"Items per page (default: 20)"
//	@Success		200
//	@Failure		401
//	@Failure		403
//	@Router			/v1/ads [get]
func (h *CMSHandler) ListAds(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	if page == 0 {
		page = 1
	}
	limit, _ := strconv.Atoi(c.Query("limit"))
	if limit == 0 {
		limit = 20
	}
	result, err := h.cmsService.ListAds(page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success(result))
}

// GetAd godoc
//
//	@Summary		Get advertisement by ID
//	@Description	Get a single advertisement by its ID. Only ADMIN role.
//	@Tags			Ads
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int	true	"Ad ID"
//	@Success		200
//	@Failure		400
//	@Failure		401
//	@Failure		403
//	@Failure		404
//	@Router			/v1/ads/{id} [get]
func (h *CMSHandler) GetAd(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("VALIDATION_ERROR", "Invalid ad ID"))
		return
	}
	ad, err := h.cmsService.GetAd(uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(
				http.StatusNotFound,
				response.Error("AD_NOT_FOUND", "Ad with ID "+c.Param("id")+" not found"),
			)
			return
		}
		c.JSON(http.StatusInternalServerError, response.Error("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success(ad))
}

// UpdateAd godoc
//
//	@Summary		Update advertisement
//	@Description	Update an existing advertisement. Only ADMIN role.
//	@Tags			Ads
//	@Accept		json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int	true	"Ad ID"
//	@Success		200
//	@Failure		400
//	@Failure		401
//	@Failure		403
//	@Failure		404
//	@Failure		422
//	@Router			/v1/ads/{id} [put]
func (h *CMSHandler) UpdateAd(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("VALIDATION_ERROR", "Invalid ad ID"))
		return
	}
	var updates model.Ad
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(
			http.StatusUnprocessableEntity,
			response.ValidationError("Invalid request body", nil),
		)
		return
	}
	ad, err := h.cmsService.UpdateAd(uint(id), &updates)
	if err != nil {
		if err.Error() == "INVALID_SCHEDULE" {
			c.JSON(
				http.StatusUnprocessableEntity,
				response.Error(
					"INVALID_SCHEDULE",
					"Start date must be before or equal to end date",
				),
			)
			return
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(
				http.StatusNotFound,
				response.Error("AD_NOT_FOUND", "Ad with ID "+c.Param("id")+" not found"),
			)
			return
		}
		c.JSON(http.StatusInternalServerError, response.Error("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success(ad))
}

// DeleteAd godoc
//
//	@Summary		Delete advertisement
//	@Description	Delete an advertisement by ID. Only ADMIN role.
//	@Tags			Ads
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	int	true	"Ad ID"
//	@Success		204
//	@Failure		400
//	@Failure		401
//	@Failure		403
//	@Failure		404
//	@Router			/v1/ads/{id} [delete]
func (h *CMSHandler) DeleteAd(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("VALIDATION_ERROR", "Invalid ad ID"))
		return
	}
	if err := h.cmsService.DeleteAd(uint(id)); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(
				http.StatusNotFound,
				response.Error("AD_NOT_FOUND", "Ad with ID "+c.Param("id")+" not found"),
			)
			return
		}
		c.JSON(http.StatusInternalServerError, response.Error("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}
	c.Status(http.StatusNoContent)
}

// ToggleAd godoc
//
//	@Summary		Toggle advertisement status
//	@Description	Activate or deactivate an advertisement. Only ADMIN role.
//	@Tags			Ads
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	int	true	"Ad ID"
//	@Success		200
//	@Failure		400
//	@Failure		401
//	@Failure		403
//	@Failure		404
//	@Router			/v1/ads/{id}/toggle [post]
func (h *CMSHandler) ToggleAd(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("VALIDATION_ERROR", "Invalid ad ID"))
		return
	}
	ad, err := h.cmsService.ToggleAd(uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(
				http.StatusNotFound,
				response.Error("AD_NOT_FOUND", "Ad with ID "+c.Param("id")+" not found"),
			)
			return
		}
		c.JSON(http.StatusInternalServerError, response.Error("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success(gin.H{
		"id":        ad.ID,
		"isActive":  ad.IsActive,
		"updatedAt": ad.UpdatedAt,
	}))
}

// ActivePlaylist godoc
//
//	@Summary		Get active playlist
//	@Description	Get list of active ads for a specific store type. Public endpoint.
//	@Tags			Ads
//	@Produce		json
//	@Param			storeType	query	string	true	"Store type (e.g., RETAIL, FOOD)"
//	@Success		200
//	@Failure		400
//	@Router			/v1/ads/active [get]
func (h *CMSHandler) ActivePlaylist(c *gin.Context) {
	storeType := c.Query("storeType")
	if storeType == "" {
		c.JSON(http.StatusBadRequest, response.Error("VALIDATION_ERROR", "storeType is required"))
		return
	}
	ads, err := h.cmsService.GetActivePlaylist(storeType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}
	type playlistItem struct {
		ID              uint   `json:"id"`
		Name            string `json:"name"`
		Filename        string `json:"filename"`
		DownloadURL     string `json:"downloadUrl"`
		PlaylistOrder   int    `json:"playlistOrder"`
		DurationSeconds *int   `json:"durationSeconds"`
	}
	playlist := make([]playlistItem, len(ads))
	for i, ad := range ads {
		playlist[i] = playlistItem{
			ID:              ad.ID,
			Name:            ad.Name,
			Filename:        ad.Filename,
			DownloadURL:     fmt.Sprintf("/v1/ads/download/%d", ad.ID),
			PlaylistOrder:   ad.PlaylistOrder,
			DurationSeconds: ad.DurationSeconds,
		}
	}
	c.JSON(http.StatusOK, response.Success(gin.H{
		"playlist":  playlist,
		"updatedAt": time.Now().UTC().Format(time.RFC3339),
	}))
}

// DownloadAd godoc
//
//	@Summary		Download advertisement video
//	@Description	Download the video file for an advertisement. Public endpoint.
//	@Tags			Ads
//	@Produce		video/mp4
//	@Param			id	path	int	true	"Ad ID"
//	@Success		200	"Video file"
//	@Failure		400
//	@Failure		404
//	@Router			/v1/ads/download/{id} [get]
func (h *CMSHandler) DownloadAd(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("VALIDATION_ERROR", "Invalid ad ID"))
		return
	}
	ad, err := h.cmsService.GetAd(uint(id))
	if err != nil {
		c.JSON(
			http.StatusNotFound,
			response.Error("AD_NOT_FOUND", "Ad with ID "+c.Param("id")+" not found"),
		)
		return
	}
	info, err := os.Stat(ad.StoragePath)
	if err != nil {
		c.JSON(http.StatusNotFound, response.Error("AD_NOT_FOUND", "File not found"))
		return
	}
	mimeType := ad.MimeType
	if mimeType == "" {
		mimeType = "video/mp4"
	}
	c.Header("Content-Type", mimeType)
	c.Header("Content-Length", strconv.FormatInt(info.Size(), 10))
	c.Header("Accept-Ranges", "bytes")
	c.Header("Cache-Control", "public, max-age=86400")
	c.Header(
		"Content-Disposition",
		fmt.Sprintf("inline; filename=\"%s\"", filepath.Base(ad.StoragePath)),
	)
	c.File(ad.StoragePath)
}

// LogPlay godoc
//
//	@Summary		Log ad play
//	@Description	Log when an ad is played on a terminal. Used for analytics.
//	@Tags			Ads
//	@Accept		json
//	@Produce		json
//	@Param			id			path		int		true	"Ad ID"
//	@Success		200
//	@Failure		400
//	@Failure		422
//	@Router			/v1/ads/{id}/play [post]
func (h *CMSHandler) LogPlay(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("VALIDATION_ERROR", "Invalid ad ID"))
		return
	}
	var req struct {
		TerminalID string `json:"terminalId"`
		StoreType  string `json:"storeType" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(
			http.StatusUnprocessableEntity,
			response.ValidationError("Invalid request body", nil),
		)
		return
	}
	if err := h.cmsService.LogPlay(uint(id), req.TerminalID, req.StoreType); err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success(gin.H{"logged": true}))
}
