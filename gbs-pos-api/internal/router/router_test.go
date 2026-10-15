package router

import (
	"net/http"
	"net/http/httptest"
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
			Discount:   handler.NewDiscountHandler(nil, nil),
			Order:      handler.NewOrderHandler(nil, nil),
			Settlement: handler.NewSettlementHandler(nil),
			Customer:   handler.NewCustomerHandler(nil),
			Dashboard:  handler.NewDashboardHandler(nil),
		},
	)

	assert.NotNil(t, r)
	for _, route := range []struct{ method, path string }{
		{http.MethodPost, "/v1/qris/payments"},
		{http.MethodGet, "/v1/qris/payments/ORDER-1/status"},
		{http.MethodPost, "/v1/btn-qris/cpm/payments"},
		{http.MethodGet, "/v1/btn-qris/cpm/payments/BTN-1"},
	} {
		request := httptest.NewRequest(route.method, route.path, nil)
		response := httptest.NewRecorder()
		r.ServeHTTP(response, request)
		assert.Equal(t, http.StatusUnauthorized, response.Code, route.path)
	}
}
