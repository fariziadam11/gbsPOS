package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"gbs-pos-api/internal/database"
	"gbs-pos-api/internal/middleware"
	"gbs-pos-api/internal/model"
	"gbs-pos-api/internal/repository"
	"gbs-pos-api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func setupHandlerTest(t *testing.T) (*gin.Engine, *service.AuthService, *service.OrderService, *service.SettlementService, *gorm.DB) {
	gin.SetMode(gin.TestMode)
	db, err := database.NewTestDB()
	require.NoError(t, err)

	os.Setenv("JWT_SECRET", "test-secret-key-minimum-32-characters")
	os.Setenv("JWT_EXPIRY_HOURS", "24")

	// Seed users
	hash, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	db.Create(&model.User{Username: "admin", PasswordHash: string(hash), Name: "Admin", Role: "ADMIN"})
	cashierHash, _ := bcrypt.GenerateFromPassword([]byte("cashier123"), bcrypt.DefaultCost)
	db.Create(&model.User{Username: "cashier", PasswordHash: string(cashierHash), Name: "Cashier", Role: "CASHIER"})

	userRepo := repository.NewUserRepository(db)
	orderRepo := repository.NewOrderRepository(db)
	settlementRepo := repository.NewSettlementRepository(db)
	productRepo := repository.NewProductRepository(db)

	authSvc := service.NewAuthService(userRepo)
	orderSvc := service.NewOrderService(orderRepo)
	settlementSvc := service.NewSettlementService(orderRepo, settlementRepo)
	productSvc := service.NewProductService(productRepo)

	authH := NewAuthHandler(authSvc)
	orderH := NewOrderHandler(orderSvc, settlementSvc)
	productH := NewProductHandler(productSvc)
	settlementH := NewSettlementHandler(settlementSvc)

	r := gin.New()
	v1 := r.Group("/v1")
	{
		v1.POST("/login", authH.Login)
		auth := v1.Group("", middleware.AuthMiddleware())
		{
			auth.GET("/products", productH.List)
			auth.POST("/products", middleware.RequireRole("ADMIN"), productH.Create)
			auth.GET("/orders", orderH.List)
			auth.POST("/orders", orderH.Create)
			auth.POST("/sync/orders", orderH.BulkSync)
			auth.PATCH("/orders/:id/void", middleware.RequireRole("ADMIN"), orderH.Void)
			auth.GET("/orders/unsettled/summary", orderH.UnsettledSummary)
			auth.POST("/orders/settle", middleware.RequireRole("ADMIN"), orderH.Settle)
			auth.GET("/settlements", settlementH.List)
		}
	}

	return r, authSvc, orderSvc, settlementSvc, db
}

func TestAuthHandler_Login(t *testing.T) {
	r, _, _, _, _ := setupHandlerTest(t)

	body, _ := json.Marshal(map[string]string{"username": "admin", "password": "admin123"})
	req := httptest.NewRequest(http.MethodPost, "/v1/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.True(t, resp["success"].(bool))
	assert.NotEmpty(t, resp["data"].(map[string]interface{})["token"])
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	r, _, _, _, _ := setupHandlerTest(t)

	body, _ := json.Marshal(map[string]string{"username": "admin", "password": "wrong"})
	req := httptest.NewRequest(http.MethodPost, "/v1/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestOrderHandler_Create(t *testing.T) {
	r, authSvc, _, _, _ := setupHandlerTest(t)

	// Login to get token
	result, _ := authSvc.Login("admin", "admin123")
	token := result.Token

	order := map[string]interface{}{
		"id":            "ORDER-001",
		"subtotal":      20000,
		"tax":           2000,
		"total":         22000,
		"paymentMethod": "CASH",
		"timestamp":     1716023456789,
		"items": []map[string]interface{}{
			{"productId": 1, "productName": "Chitato", "productPrice": 10000, "qty": 2, "subtotal": 20000},
		},
	}
	body, _ := json.Marshal(order)
	req := httptest.NewRequest(http.MethodPost, "/v1/orders", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.True(t, resp["success"].(bool))
}

func TestOrderHandler_Create_Idempotent(t *testing.T) {
	r, authSvc, _, _, _ := setupHandlerTest(t)
	result, _ := authSvc.Login("admin", "admin123")
	token := result.Token

	order := map[string]interface{}{
		"id":            "ORDER-001",
		"subtotal":      20000,
		"tax":           2000,
		"total":         22000,
		"paymentMethod": "CASH",
		"timestamp":     1716023456789,
		"items":         []map[string]interface{}{{"productId": 1, "productName": "Chitato", "productPrice": 10000, "qty": 2, "subtotal": 20000}},
	}

	// First create
	body, _ := json.Marshal(order)
	req1 := httptest.NewRequest(http.MethodPost, "/v1/orders", bytes.NewReader(body))
	req1.Header.Set("Content-Type", "application/json")
	req1.Header.Set("Authorization", "Bearer "+token)
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusCreated, w1.Code)

	// Second create (idempotent)
	req2 := httptest.NewRequest(http.MethodPost, "/v1/orders", bytes.NewReader(body))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Authorization", "Bearer "+token)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)

	var resp map[string]interface{}
	json.Unmarshal(w2.Body.Bytes(), &resp)
	assert.True(t, resp["idempotent"].(bool))
}

func TestOrderHandler_Void_ForbiddenForCashier(t *testing.T) {
	r, authSvc, _, _, _ := setupHandlerTest(t)
	result, _ := authSvc.Login("cashier", "cashier123")
	token := result.Token

	req := httptest.NewRequest(http.MethodPatch, "/v1/orders/ORDER-001/void", bytes.NewReader([]byte(`{"reason":"test"}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestOrderHandler_Void_AdminSuccess(t *testing.T) {
	r, authSvc, _, _, db := setupHandlerTest(t)
	result, _ := authSvc.Login("admin", "admin123")
	token := result.Token

	// Create order
	db.Create(&model.Order{ID: "ORDER-VOID", Total: 10000, PaymentMethod: "CASH", Timestamp: 1716023456789, IsVoided: false, IsSettled: false})

	req := httptest.NewRequest(http.MethodPatch, "/v1/orders/ORDER-VOID/void", bytes.NewReader([]byte(`{"reason":"Customer requested"}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestProductHandler_Create_ForbiddenForCashier(t *testing.T) {
	r, authSvc, _, _, _ := setupHandlerTest(t)
	result, _ := authSvc.Login("cashier", "cashier123")
	token := result.Token

	product := map[string]interface{}{"name": "Test", "price": 1000, "category": "Test", "storeType": "RETAIL"}
	body, _ := json.Marshal(product)
	req := httptest.NewRequest(http.MethodPost, "/v1/products", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}
