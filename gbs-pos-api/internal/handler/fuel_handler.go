package handler

import (
	"gbs-common/pkg/response"
	"gbs-pos-api/internal/dto"
	"gbs-pos-api/internal/service"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type FuelHandler struct {
	fuelService *service.FuelService
}

func NewFuelHandler(fuelService *service.FuelService) *FuelHandler {
	return &FuelHandler{fuelService: fuelService}
}

// Fuel prices
func (h *FuelHandler) ListPrices(c *gin.Context) {
	prices, err := h.fuelService.ListPrices()
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success(prices))
}

func (h *FuelHandler) UpdatePrice(c *gin.Context) {
	code := c.Param("code")
	var req dto.UpdateFuelPriceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, response.ValidationError("Invalid request body", nil))
		return
	}
	price, err := h.fuelService.UpdatePrice(code, req)
	if err != nil {
		if service.IsFuelNotFound(err) {
			c.JSON(http.StatusNotFound, response.Error("NOT_FOUND", "Fuel price not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, response.Error("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success(price))
}

// Pumps
func (h *FuelHandler) ListPumps(c *gin.Context) {
	pumps, err := h.fuelService.ListPumps()
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success(pumps))
}

func (h *FuelHandler) CreatePump(c *gin.Context) {
	var req dto.CreatePumpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, response.ValidationError("Invalid request body", nil))
		return
	}
	pump, err := h.fuelService.CreatePump(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}
	c.JSON(http.StatusCreated, response.Success(pump))
}

func (h *FuelHandler) UpdatePump(c *gin.Context) {
	id := c.Param("id")
	var req dto.UpdatePumpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, response.ValidationError("Invalid request body", nil))
		return
	}
	pump, err := h.fuelService.UpdatePump(id, req)
	if err != nil {
		if service.IsFuelNotFound(err) {
			c.JSON(http.StatusNotFound, response.Error("NOT_FOUND", "Pump not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, response.Error("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success(pump))
}

func (h *FuelHandler) DeletePump(c *gin.Context) {
	id := c.Param("id")
	if err := h.fuelService.DeletePump(id); err != nil {
		if service.IsFuelNotFound(err) {
			c.JSON(http.StatusNotFound, response.Error("NOT_FOUND", "Pump not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, response.Error("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}
	c.Status(http.StatusNoContent)
}

// Nozzles
func (h *FuelHandler) ListNozzles(c *gin.Context) {
	nozzles, err := h.fuelService.ListNozzles()
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success(nozzles))
}

func (h *FuelHandler) CreateNozzle(c *gin.Context) {
	var req dto.CreateNozzleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, response.ValidationError("Invalid request body", nil))
		return
	}
	nozzle, err := h.fuelService.CreateNozzle(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}
	c.JSON(http.StatusCreated, response.Success(nozzle))
}

func (h *FuelHandler) UpdateNozzle(c *gin.Context) {
	id := c.Param("id")
	var req dto.UpdateNozzleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, response.ValidationError("Invalid request body", nil))
		return
	}
	nozzle, err := h.fuelService.UpdateNozzle(id, req)
	if err != nil {
		if service.IsFuelNotFound(err) {
			c.JSON(http.StatusNotFound, response.Error("NOT_FOUND", "Nozzle not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, response.Error("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success(nozzle))
}

func (h *FuelHandler) DeleteNozzle(c *gin.Context) {
	id := c.Param("id")
	if err := h.fuelService.DeleteNozzle(id); err != nil {
		if service.IsFuelNotFound(err) {
			c.JSON(http.StatusNotFound, response.Error("NOT_FOUND", "Nozzle not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, response.Error("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}
	c.Status(http.StatusNoContent)
}

// Fuel sales
func (h *FuelHandler) CreateSale(c *gin.Context) {
	var req dto.FuelSaleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, response.ValidationError("Invalid request body", nil))
		return
	}
	sale, err := h.fuelService.CreateSale(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}
	c.JSON(http.StatusCreated, response.Success(sale))
}

func (h *FuelHandler) Report(c *gin.Context) {
	fromStr := c.Query("from")
	toStr := c.Query("to")
	from, err := time.Parse(time.DateOnly, fromStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("VALIDATION_ERROR", "Invalid from date"))
		return
	}
	to, err := time.Parse(time.DateOnly, toStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("VALIDATION_ERROR", "Invalid to date"))
		return
	}
	to = to.Add(24*time.Hour - time.Second)
	report, err := h.fuelService.Report(from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success(report))
}
