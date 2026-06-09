package router

import (
	"testing"

	"gbs-pos-api/internal/config"
	"gbs-pos-api/internal/handler"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSetupRegistersRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := Setup(
		&config.Config{JWTSecret: "test-secret-key-minimum-32-characters"},
		Handlers{
			Auth:       handler.NewAuthHandler(nil),
			Product:    handler.NewProductHandler(nil),
			Discount:   handler.NewDiscountHandler(nil),
			Order:      handler.NewOrderHandler(nil, nil),
			Settlement: handler.NewSettlementHandler(nil),
			Customer:   handler.NewCustomerHandler(nil),
			Dashboard:  handler.NewDashboardHandler(nil),
		},
	)

	assert.NotNil(t, r)
}
