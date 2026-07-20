package service

import (
	"context"
	"testing"
	"time"

	"gbs-pos-api/internal/config"
	"gbs-pos-api/internal/dto"
	"gbs-pos-api/internal/model"
	"gbs-pos-api/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupQrisTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&model.Order{})
	require.NoError(t, err)

	return db
}

func setupTestConfig() *config.Config {
	return &config.Config{
		SumoPodAPIURL:        "https://api-pay-sandbox.sumopod.com/api/v1",
		SumoPodAPIKey:        "test-api-key",
		SumoPodWebhookToken:  "test-webhook-token",
		SumoPodExpiresHours: 24,
	}
}

func TestQrisService_InitPayment_OrderNotFound(t *testing.T) {
	db := setupQrisTestDB(t)
	cfg := setupTestConfig()

	orderRepo := repository.NewOrderRepository(db)
	qrisService := NewQrisService(cfg, db, orderRepo)

	result, err := qrisService.InitPayment(context.Background(), dto.CreateQrisPaymentRequest{
		OrderID: "NONEXISTENT",
		Amount:  50000,
	})
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "ORDER_NOT_FOUND")
}

func TestQrisService_InitPayment_NotQrisPayment(t *testing.T) {
	db := setupQrisTestDB(t)
	cfg := setupTestConfig()

	order := &model.Order{
		ID:            "TEST-CASH-001",
		Total:         50000,
		PaymentMethod: "CASH",
	}
	db.Create(order)

	orderRepo := repository.NewOrderRepository(db)
	qrisService := NewQrisService(cfg, db, orderRepo)

	result, err := qrisService.InitPayment(context.Background(), dto.CreateQrisPaymentRequest{
		OrderID: "TEST-CASH-001",
		Amount:  50000,
	})
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "INVALID_PAYMENT_METHOD")
}

func TestQrisService_GetPaymentStatus_Success(t *testing.T) {
	db := setupQrisTestDB(t)
	cfg := setupTestConfig()

	expiresAt := time.Now().Add(24 * time.Hour)
	completedAt := time.Now()

	order := &model.Order{
		ID:              "TEST-002",
		Total:           50000,
		PaymentMethod:   "QRIS",
		QrisPaymentID:   "PAY-123",
		QrisStatus:      dto.QrisStatusCompleted,
		QrisLinkURL:     "https://payment.link",
		QrisExpiresAt:   &expiresAt,
		QrisFee:         750,
		QrisNetAmount:   49250,
		QrisCompletedAt: &completedAt,
	}
	db.Create(order)

	orderRepo := repository.NewOrderRepository(db)
	qrisService := NewQrisService(cfg, db, orderRepo)

	status, err := qrisService.GetPaymentStatus(context.Background(), "TEST-002")
	assert.NoError(t, err)
	assert.NotNil(t, status)
	assert.Equal(t, "TEST-002", status.OrderID)
	assert.Equal(t, "PAY-123", status.PaymentID)
	assert.Equal(t, dto.QrisStatusCompleted, status.Status)
	assert.Equal(t, 750.0, status.Fee)
	assert.Equal(t, 49250.0, status.NetAmount)
}

func TestQrisService_HandleWebhook_TestEvent(t *testing.T) {
	db := setupQrisTestDB(t)
	cfg := setupTestConfig()

	qrisService := NewQrisService(cfg, db, nil)

	payload := dto.SumoPodWebhookPayload{
		EventType: dto.WebhookEventPaymentTest,
		Data: dto.SumoPodWebhookPaymentData{
			PaymentID: "PAY-TEST",
			OrderID:   "TEST-ORDER",
		},
	}

	err := qrisService.HandleWebhook(context.Background(), payload)
	assert.NoError(t, err)
}

func TestQrisService_HandleWebhook_Completed(t *testing.T) {
	db := setupQrisTestDB(t)
	cfg := setupTestConfig()

	order := &model.Order{
		ID:             "TEST-003",
		Total:          50000,
		PaymentMethod:  "QRIS",
		QrisPaymentID:  "PAY-456",
		QrisStatus:     dto.QrisStatusPending,
	}
	db.Create(order)

	orderRepo := repository.NewOrderRepository(db)
	qrisService := NewQrisService(cfg, db, orderRepo)

	completedAt := time.Now()
	payload := dto.SumoPodWebhookPayload{
		EventType: dto.WebhookEventPaymentCompleted,
		Data: dto.SumoPodWebhookPaymentData{
			PaymentID:   "PAY-456",
			OrderID:     "TEST-003",
			Amount:      50000,
			Fee:         750,
			NetAmount:   49250,
			Status:      "completed",
			CompletedAt: completedAt,
		},
	}

	err := qrisService.HandleWebhook(context.Background(), payload)
	assert.NoError(t, err)

	var updatedOrder model.Order
	db.Where("id = ?", "TEST-003").First(&updatedOrder)
	assert.Equal(t, dto.QrisStatusCompleted, updatedOrder.QrisStatus)
	assert.NotNil(t, updatedOrder.QrisCompletedAt)
	assert.Equal(t, 750.0, updatedOrder.QrisFee)
	assert.Equal(t, 49250.0, updatedOrder.QrisNetAmount)
}

func TestQrisService_HandleWebhook_Failed(t *testing.T) {
	db := setupQrisTestDB(t)
	cfg := setupTestConfig()

	order := &model.Order{
		ID:             "TEST-004",
		Total:          50000,
		PaymentMethod:  "QRIS",
		QrisPaymentID:  "PAY-789",
		QrisStatus:     dto.QrisStatusPending,
	}
	db.Create(order)

	orderRepo := repository.NewOrderRepository(db)
	qrisService := NewQrisService(cfg, db, orderRepo)

	payload := dto.SumoPodWebhookPayload{
		EventType: dto.WebhookEventPaymentFailed,
		Data: dto.SumoPodWebhookPaymentData{
			PaymentID: "PAY-789",
			OrderID:   "TEST-004",
		},
	}

	err := qrisService.HandleWebhook(context.Background(), payload)
	assert.NoError(t, err)

	var updatedOrder model.Order
	db.Where("id = ?", "TEST-004").First(&updatedOrder)
	assert.Equal(t, dto.QrisStatusFailed, updatedOrder.QrisStatus)
}

func TestQrisService_HandleWebhook_Expired(t *testing.T) {
	db := setupQrisTestDB(t)
	cfg := setupTestConfig()

	order := &model.Order{
		ID:             "TEST-005",
		Total:          50000,
		PaymentMethod:  "QRIS",
		QrisPaymentID:  "PAY-EXP",
		QrisStatus:     dto.QrisStatusPending,
	}
	db.Create(order)

	orderRepo := repository.NewOrderRepository(db)
	qrisService := NewQrisService(cfg, db, orderRepo)

	payload := dto.SumoPodWebhookPayload{
		EventType: dto.WebhookEventPaymentExpired,
		Data: dto.SumoPodWebhookPaymentData{
			PaymentID: "PAY-EXP",
			OrderID:   "TEST-005",
		},
	}

	err := qrisService.HandleWebhook(context.Background(), payload)
	assert.NoError(t, err)

	var updatedOrder model.Order
	db.Where("id = ?", "TEST-005").First(&updatedOrder)
	assert.Equal(t, dto.QrisStatusExpired, updatedOrder.QrisStatus)
}

func TestGenerateOrderID(t *testing.T) {
	orderID := GenerateOrderID()
	assert.Contains(t, orderID, "QRIS-")
	assert.Len(t, orderID, 17)
}
