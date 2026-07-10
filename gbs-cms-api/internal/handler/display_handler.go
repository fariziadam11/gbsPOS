package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"gbs-cms-api/internal/dto"
	"gbs-cms-api/internal/service"
	"gbs-common/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Default cart display JSON when no data exists
const DefaultCartDisplayJSON = `{
  "Initial": {
    "NamaKasir": "DOMAR",
    "KodeToko": "T14AB",
    "NamaToko": "Indomaret Pusat",
    "JenisToko": "POINT"
  },
  "DaftarBelanja": [],
  "Summary": {
    "Hemat": "0",
    "Total": "0",
    "Bayar": "0",
    "Kembali": "0"
  },
  "TeksSelesai": "None"
}`

type DisplayHandler struct {
	service *service.CartDisplayService
}

func NewDisplayHandler(service *service.CartDisplayService) *DisplayHandler {
	return &DisplayHandler{service: service}
}

// SaveCartDisplay godoc
//
//	@Summary		Upload cart display JSON
//	@Description	Receives cart display JSON from Android POS and stores it. Requires Bearer token authentication.
//	@Tags			Cart Display
//	@Accept		json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body	string	true	"Cart display JSON (must include terminalId field)"
//	@Success		200
//	@Failure		400
//	@Failure		401
//	@Router			/v1/display/cart [post]
func (h *DisplayHandler) SaveCartDisplay(c *gin.Context) {
	// Read entire body as string
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest,
			response.Error("BAD_REQUEST", "Failed to read request body"))
		return
	}

	payload := string(body)
	if payload == "" {
		c.JSON(http.StatusBadRequest,
			response.Error("BAD_REQUEST", "Empty payload"))
		return
	}

	// Extract terminalId and deviceInfo from the JSON payload
	terminalID, deviceInfo := extractPayloadData(payload)
	if terminalID == "" {
		c.JSON(http.StatusBadRequest,
			response.Error("VALIDATION_ERROR", "terminalId is required"))
		return
	}

	if err := h.service.SaveCartDisplay(terminalID, payload, deviceInfo); err != nil {
		c.JSON(http.StatusInternalServerError,
			response.Error("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}

// GetCartDisplay godoc
//
//	@Summary		Get cart display JSON
//	@Description	Returns the latest stored cart display JSON for a terminal. Public endpoint - no authentication required.
//	@Tags			Cart Display
//	@Produce		json
//	@Param			terminalId	query	string	false	"Android terminal ID (ANDROID_ID)"
//	@Success		200	{string}	string	"JSON cart display data"
//	@Failure		500
//	@Router			/v1/display/cart [get]
func (h *DisplayHandler) GetCartDisplay(c *gin.Context) {
	terminalID := c.Query("terminalId")
	if terminalID == "" {
		c.Data(http.StatusOK, "application/json", []byte(DefaultCartDisplayJSON))
		return
	}

	payload, err := h.service.GetCartDisplay(terminalID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.Data(http.StatusOK, "application/json", []byte(DefaultCartDisplayJSON))
			return
		}
		c.JSON(http.StatusInternalServerError,
			response.Error("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}

	c.Data(http.StatusOK, "application/json", []byte(payload))
}

// ListTerminals godoc
//
//	@Summary		List all terminals
//	@Description	Returns all stored cart displays with device info. Requires authentication.
//	@Tags			Cart Display
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200
//	@Failure		401
//	@Failure		500
//	@Router			/v1/display/terminals [get]
func (h *DisplayHandler) ListTerminals(c *gin.Context) {
	terminals, err := h.service.GetAllTerminals()
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			response.Error("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(terminals))
}

// extractPayloadData extracts terminalId and deviceInfo from JSON payload
func extractPayloadData(jsonStr string) (string, *dto.DeviceInfo) {
	// Parse the JSON to extract terminalId and deviceInfo
	var raw map[string]json.RawMessage
	if err := json.Unmarshal([]byte(jsonStr), &raw); err != nil {
		return "", nil
	}

	// Extract terminalId
	var terminalID string
	if terminalIDRaw, ok := raw["terminalId"]; ok {
		json.Unmarshal(terminalIDRaw, &terminalID)
	}
	if terminalID == "" {
		return "", nil
	}

	// Extract deviceInfo (optional)
	var deviceInfo *dto.DeviceInfo
	if deviceInfoRaw, ok := raw["deviceInfo"]; ok && deviceInfoRaw != nil {
		var di dto.DeviceInfo
		if err := json.Unmarshal(deviceInfoRaw, &di); err == nil {
			deviceInfo = &di
		}
	}

	return terminalID, deviceInfo
}

// extractTerminalID extracts terminalId from JSON payload (legacy helper)
func extractTerminalID(json string) string {
	prefix := `"terminalId"`
	idx := strings.Index(json, prefix)
	if idx == -1 {
		return ""
	}

	start := idx + len(prefix)
	for start < len(json) && (json[start] == ' ' || json[start] == ':' || json[start] == '\t') {
		start++
	}
	if start >= len(json) {
		return ""
	}

	if json[start] == '"' {
		start++
		end := start
		for end < len(json) && json[end] != '"' {
			end++
		}
		return json[start:end]
	}

	end := start
	for end < len(json) && json[end] != ',' && json[end] != '}' && json[end] != ' ' && json[end] != '\n' && json[end] != '\t' {
		end++
	}
	return json[start:end]
}
