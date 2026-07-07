package handler

import (
	"io"
	"net/http"
	"strings"

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

	// Extract terminalId from the JSON payload
	terminalID := extractTerminalID(payload)
	if terminalID == "" {
		c.JSON(http.StatusBadRequest,
			response.Error("VALIDATION_ERROR", "terminalId is required"))
		return
	}

	if err := h.service.SaveCartDisplay(terminalID, payload); err != nil {
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

// extractTerminalID extracts terminalId from JSON payload
func extractTerminalID(json string) string {
	// Simple extraction - find "terminalId" and get the value
	prefix := `"terminalId"`
	idx := strings.Index(json, prefix)
	if idx == -1 {
		return ""
	}

	// Find the value after the key
	start := idx + len(prefix)
	// Skip whitespace and colon
	for start < len(json) && (json[start] == ' ' || json[start] == ':' || json[start] == '\t') {
		start++
	}
	if start >= len(json) {
		return ""
	}

	// Check if value is quoted
	if json[start] == '"' {
		// Quoted string value
		start++ // skip opening quote
		end := start
		for end < len(json) && json[end] != '"' {
			end++
		}
		return json[start:end]
	}

	// Unquoted value (simple case)
	end := start
	for end < len(json) && json[end] != ',' && json[end] != '}' && json[end] != ' ' && json[end] != '\n' && json[end] != '\t' {
		end++
	}
	return json[start:end]
}
