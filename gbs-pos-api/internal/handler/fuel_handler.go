package handler

import (
	"gbs-pos-api/internal/dto"
	"gbs-pos-api/internal/service"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"gbs-common/pkg/response"
)

type FuelHandler struct {
	fuelService *service.FuelService
}

func NewFuelHandler(fuelService *service.FuelService) *FuelHandler {
	return &FuelHandler{fuelService: fuelService}
}

// ListPrices godoc
//
//	@Summary		List fuel prices
//	@Description	Get all fuel prices
//	@Tags			Fuel
//	@Produce		json
//	@Success		200
//	@Router				/v1/fuel/prices [get]
func (h *FuelHandler) ListPrices(c *gin.Context) {
	prices, err := h.fuelService.ListPrices()
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success(prices))
}

// UpdatePrice godoc
//
//	@Summary		Update fuel price
//	@Description	Update fuel price by code
//	@Tags			Fuel
//	@Accept		json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			code	path	string					true	"Fuel code (PERTALITE, PERTAMAX, DEXlite, etc)"
//	@Param			request	body	dto.UpdateFuelPriceRequest	true	"Price update data"
//	@Success		200
//	@Failure		400
//	@Failure		404
//	@Failure		422
//	@Router				/v1/fuel/prices/{code} [patch]
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

// ListPumps godoc
//
//	@Summary		List pumps
//	@Description	Get all pumps
//	@Tags			Fuel
//	@Produce		json
//	@Success		200
//	@Failure		500
//	@Router				/v1/fuel/pumps [get]
func (h *FuelHandler) ListPumps(c *gin.Context) {
	pumps, err := h.fuelService.ListPumps()
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success(pumps))
}

// CreatePump godoc
//
//	@Summary		Create pump
//	@Description	Create a new pump
//	@Tags			Fuel
//	@Accept		json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body	dto.CreatePumpRequest	true	"Pump data"
//	@Success		201
//	@Failure		401
//	@Failure		422
//	@Router				/v1/fuel/pumps [post]
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

// UpdatePump godoc
//
//	@Summary		Update pump
//	@Description	Update pump details
//	@Tags			Fuel
//	@Accept		json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path	string				true	"Pump ID"
//	@Param			request	body	dto.UpdatePumpRequest	true	"Update data"
//	@Success		200
//	@Failure		400
//	@Failure		401
//	@Failure		404
//	@Failure		422
//	@Router				/v1/fuel/pumps/{id} [patch]
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

// DeletePump godoc
//
//	@Summary		Delete pump
//	@Description	Delete a pump
//	@Tags			Fuel
//	@Security		BearerAuth
//	@Param			id	path	string	true	"Pump ID"
//	@Success		204
//	@Failure		401
//	@Failure		404
//	@Router				/v1/fuel/pumps/{id} [delete]
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

// ListNozzles godoc
//
//	@Summary		List nozzles
//	@Description	Get all nozzles
//	@Tags			Fuel
//	@Produce		json
//	@Success		200
//	Failure			401
//	@Router				/v1/fuel/nozzles [get]
func (h *FuelHandler) ListNozzles(c *gin.Context) {
	nozzles, err := h.fuelService.ListNozzles()
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success(nozzles))
}

// CreateNozzle godoc
//
//	@Summary		Create nozzle
//	@Description	Create a new nozzle
//	@Tags			Fuel
//	@Accept		json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body	dto.CreateNozzleRequest	true	"Nozzle data"
//	@Success		201
//	@Failure		401
//	@Failure		422
//	@Router				/v1/fuel/nozzles [post]
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

// UpdateNozzle godoc
//
//	@Summary		Update nozzle
//	@Description	Update nozzle
//	@Tags			Fuel
//	@Accept		json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path	string					true	"Nozzle ID"
//	@Param			request	body	dto.UpdateNozzleRequest	true	"Update data"
//	@Success		200
//	@Failure		400
//	@Failure		401
//	@Failure		404
//	@Failure		422
//	@Router				/v1/fuel/nozzles/{id} [patch]
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

// DeleteNozzle godoc
//
//	@Summary		Delete nozzle
//	@Description	Delete a nozzle
//	@Tags			Fuel
//	@Security		BearerAuth
//	@Param			id	path	string	true	"Nozzle ID"
//	@Success		204
//	@Failure		401
//	@Failure		404
//	@Router				/v1/fuel/nozzles/{id} [delete]
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

// CreateSale godoc
//
//	@Summary		Record fuel sale
//	@Description	Record a fuel sale transaction
//	@Tags			Fuel
//	@Accept		json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body	dto.FuelSaleRequest	true	"Sale data"
//	@Success		201
//	@Failure		401
//	@Failure		422
//	@Router				/v1/fuel/sales [post]
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

// Report godoc
//
//	@Summary		Fuel sales report
//	@Description	Get fuel sales report for date range
//	@Tags			Fuel
//	@Produce		json
//	@Security		BearerAuth
//	@Param			from	query	string	true	"Start date (YYYY-MM-DD)"
//	@Param			to		query	string	true	"End date (YYYY-MM-DD)"
//	@Success		200
//	@Failure		400
//	@Failure		401
//	@Failure		500
//	@Router				/v1/fuel/report [get]
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
