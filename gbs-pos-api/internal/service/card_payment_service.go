package service

import (
	"context"
	"fmt"
	"math"
	"time"

	"gbs-pos-api/internal/model"
	"gbs-pos-api/internal/repository"
	ws "gbs-pos-api/internal/websocket"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type CardPaymentService struct {
	repo    *repository.CardPaymentRepository
	orders  *OrderService
	devices *repository.CompanionDeviceRepository
	hub     *ws.Hub
}

func NewCardPaymentService(repo *repository.CardPaymentRepository, orders *OrderService, devices *repository.CompanionDeviceRepository, hub *ws.Hub) *CardPaymentService {
	return &CardPaymentService{repo: repo, orders: orders, devices: devices, hub: hub}
}

func (s *CardPaymentService) Create(ctx context.Context, orderID string, amount float64, terminalID, deviceID string) (*model.CardPayment, error) {
	if orderID == "" || amount <= 0 || terminalID == "" || deviceID == "" {
		return nil, fmt.Errorf("orderId, amount, terminalId, and deviceId are required")
	}
	order, err := s.orders.Get(orderID)
	if err != nil {
		return nil, fmt.Errorf("order not found")
	}
	if order.IsVoided || order.IsSettled {
		return nil, fmt.Errorf("order is not payable")
	}
	if math.Abs(order.Total-amount) > 0.01 {
		return nil, fmt.Errorf("payment amount does not match order total")
	}
	if order.TerminalID != "" && order.TerminalID != terminalID {
		return nil, fmt.Errorf("terminal does not match order")
	}
	registered, err := s.devices.Exists(ctx, deviceID)
	if err != nil {
		return nil, err
	}
	if !registered {
		return nil, fmt.Errorf("companion device is not registered")
	}
	payment := &model.CardPayment{
		ID:         uuid.New(),
		OrderID:    orderID,
		Amount:     amount,
		Status:     model.CardPaymentWaiting,
		TerminalID: terminalID,
		DeviceID:   deviceID,
		ExpiresAt:  time.Now().Add(model.CardPaymentExpiryDuration),
	}
	if err := s.repo.Create(ctx, payment); err != nil {
		return nil, err
	}
	s.broadcastRequest(payment)
	s.broadcastStatus(payment)
	return payment, nil
}

func (s *CardPaymentService) Get(ctx context.Context, id uuid.UUID) (*model.CardPayment, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *CardPaymentService) Pending(ctx context.Context, deviceID string) ([]model.CardPayment, error) {
	if deviceID == "" {
		return nil, fmt.Errorf("deviceId is required")
	}
	return s.repo.FindPendingByDevice(ctx, deviceID)
}

func (s *CardPaymentService) Cancel(ctx context.Context, id uuid.UUID) (*model.CardPayment, error) {
	// Atomic: only cancels while the payment is still WAITING_FOR_CARD. If a SUCCESS
	// from the companion landed first, this returns an error instead of voiding an
	// order that may already have been paid.
	updated, err := s.repo.CancelIfWaiting(ctx, id)
	if err != nil {
		return nil, err
	}
	if !updated {
		return nil, fmt.Errorf("payment cannot be cancelled in current status")
	}
	// Void the order even if the re-fetch fails — the payment is already CANCELLED
	// and must not leave the order locked.
	payment, findErr := s.repo.FindByID(ctx, id)
	if findErr == nil {
		s.voidOrder(payment.OrderID, "card payment cancelled")
		s.broadcastStatus(payment)
		return payment, nil
	}
	return nil, findErr
}

func (s *CardPaymentService) UpdateFromCompanion(ctx context.Context, client *ws.Client, message ws.Message) error {
	if client.Type != ws.ClientCompanion || message.PaymentID == "" {
		return fmt.Errorf("invalid companion payment update")
	}
	id, err := uuid.Parse(message.PaymentID)
	if err != nil {
		return fmt.Errorf("invalid payment id")
	}
	payment, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if payment.DeviceID != client.ID {
		return fmt.Errorf("payment does not belong to companion")
	}
	if message.Status != model.CardPaymentProcessing && message.Status != model.CardPaymentSuccess && message.Status != model.CardPaymentFailed {
		return fmt.Errorf("invalid payment status")
	}
	if message.Status == model.CardPaymentSuccess && message.TransactionID == "" {
		return fmt.Errorf("successful payment requires transaction id")
	}

	// Duplicate terminal update from the companion — silently ignore.
	if payment.Status == model.CardPaymentSuccess {
		return nil
	}
	// A late SUCCESS after the payment was expired or cancelled means money may have
	// been charged on a SoftPOS while the order was (or is being) voided. Log loudly
	// and notify the POS so staff can reconcile/refund — never hide it.
	if payment.Status == model.CardPaymentExpired || payment.Status == model.CardPaymentCancelled {
		if message.Status == model.CardPaymentSuccess {
			log.Error().
				Str("payment_id", payment.ID.String()).
				Str("order_id", payment.OrderID).
				Str("payment_status", payment.Status).
				Str("transaction_id", message.TransactionID).
				Msg("LATE card payment success after payment was " + payment.Status + " — verify charge & refund if needed")
			s.hub.Send(ws.ClientPOS, payment.TerminalID, ws.Message{
				Type:          "PAYMENT_LATE_SUCCESS",
				PaymentID:     payment.ID.String(),
				OrderID:       payment.OrderID,
				Status:        message.Status,
				TransactionID: message.TransactionID,
				CardBrand:     message.CardBrand,
				MaskedCard:    message.MaskedCard,
				AuthCode:      message.AuthCode,
				EntryMode:     message.EntryMode,
				AcqMID:        message.AcqMID,
				AcqTID:        message.AcqTID,
				PosMessageID:  message.PosMessageID,
				FailureReason: "payment " + payment.Status + " before success arrived — verify charge and refund if needed",
			})
		}
		return nil
	}
	if payment.Status == model.CardPaymentFailed {
		return nil
	}
	if message.Status == model.CardPaymentProcessing {
		if payment.Status != model.CardPaymentWaiting {
			// Already PROCESSING (e.g. reconnect re-send) — idempotent, no expiry refresh.
			return nil
		}
		// Atomic WAITING→PROCESSING with a refreshed expiry window so a slow
		// SoftPOS transaction is not cut off by the expiry job. Loses the race
		// (Cancel/Expire won) when this returns false.
		newExpiry := time.Now().Add(model.CardPaymentExpiryDuration)
		updated, err := s.repo.MarkProcessing(ctx, id, newExpiry)
		if err != nil {
			return err
		}
		if !updated {
			return fmt.Errorf("payment is no longer waiting for card")
		}
		payment.Status = model.CardPaymentProcessing
		payment.ExpiresAt = newExpiry
		payment.UpdatedAt = time.Now()
		payment.PosMessageID = message.PosMessageID
		s.broadcastStatus(payment)
		return nil
	}

	if payment.Status != model.CardPaymentWaiting && payment.Status != model.CardPaymentProcessing {
		return fmt.Errorf("payment cannot complete from status %s", payment.Status)
	}
	payment.Status = message.Status
	payment.TransactionID = message.TransactionID
	payment.CardBrand = message.CardBrand
	payment.MaskedCard = message.MaskedCard
	payment.AuthCode = message.AuthCode
	payment.EntryMode = message.EntryMode
	payment.AcqMID = message.AcqMID
	payment.AcqTID = message.AcqTID
	payment.PosMessageID = message.PosMessageID
	payment.FailureReason = message.FailureReason
	payment.UpdatedAt = time.Now()
	if payment.Status == model.CardPaymentSuccess {
		if err := s.repo.FinalizeSuccess(ctx, payment); err != nil {
			return err
		}
	} else {
		if err := s.repo.Update(ctx, payment); err != nil {
			return err
		}
	}
	if payment.Status == model.CardPaymentFailed {
		s.voidOrder(payment.OrderID, "card payment not completed")
	}
	s.broadcastStatus(payment)
	return nil
}

func (s *CardPaymentService) Expire(ctx context.Context) error {
	payments, err := s.repo.FindExpired(ctx, time.Now())
	if err != nil {
		return err
	}
	for _, payment := range payments {
		// Only WAITING payments get their order voided: nothing has been charged yet.
		// A PROCESSING payment that expired (e.g. companion died mid-transaction) is
		// marked EXPIRED but its order is left open so staff can reconcile whether the
		// SoftPOS actually charged the card.
		wasProcessing := payment.Status == model.CardPaymentProcessing
		updated, err := s.repo.MarkExpired(ctx, payment.ID, time.Now())
		if err != nil {
			return err
		}
		if updated {
			payment.Status = model.CardPaymentExpired
			payment.UpdatedAt = time.Now()
			if !wasProcessing {
				s.voidOrder(payment.OrderID, "card payment expired")
			}
			s.broadcastStatus(&payment)
		}
	}
	return nil
}

func (s *CardPaymentService) voidOrder(orderID, reason string) {
	if _, err := s.orders.Void(orderID, reason, "system"); err != nil {
		log.Error().Err(err).Str("order_id", orderID).Str("reason", reason).Msg("failed to void card payment order")
	}
}

func (s *CardPaymentService) HandleMessage(client *ws.Client, message ws.Message) {
	if message.Type == "PING" {
		s.hub.Send(client.Type, client.ID, ws.Message{Type: "PONG"})
		return
	}
	if message.Type == "PAYMENT_STATUS_UPDATE" {
		if err := s.UpdateFromCompanion(context.Background(), client, message); err != nil {
			log.Error().
				Err(err).
				Str("client", client.Type+":"+client.ID).
				Str("payment_id", message.PaymentID).
				Str("status", message.Status).
				Msg("companion payment update failed")
			s.hub.Send(client.Type, client.ID, ws.Message{
				Type:          "PAYMENT_UPDATE_ERROR",
				PaymentID:     message.PaymentID,
				Status:        message.Status,
				FailureReason: err.Error(),
			})
		}
	}
}

func (s *CardPaymentService) broadcastRequest(payment *model.CardPayment) {
	s.hub.Send(ws.ClientCompanion, payment.DeviceID, ws.Message{
		Type:       "PAYMENT_REQUEST",
		PaymentID:  payment.ID.String(),
		OrderID:    payment.OrderID,
		Amount:     payment.Amount,
		Currency:   "IDR",
		ExpiresAt:  payment.ExpiresAt.UTC().Format(time.RFC3339),
		TerminalID: payment.TerminalID,
	})
}

func (s *CardPaymentService) broadcastStatus(payment *model.CardPayment) {
	s.hub.Send(ws.ClientPOS, payment.TerminalID, ws.Message{
		Type:          "PAYMENT_STATUS",
		PaymentID:     payment.ID.String(),
		OrderID:       payment.OrderID,
		Status:        payment.Status,
		Amount:        payment.Amount,
		TransactionID: payment.TransactionID,
		CardBrand:     payment.CardBrand,
		MaskedCard:    payment.MaskedCard,
		AuthCode:      payment.AuthCode,
		EntryMode:     payment.EntryMode,
		AcqMID:        payment.AcqMID,
		AcqTID:        payment.AcqTID,
		PosMessageID:  payment.PosMessageID,
		FailureReason: payment.FailureReason,
		UpdatedAt:     payment.UpdatedAt.UTC().Format(time.RFC3339),
	})
}
