package service

import (
	"fmt"
	"time"
	"gbs-pos-api/internal/model"
	"gbs-pos-api/internal/repository"
	"gorm.io/gorm"
)

type SettlementService struct {
	orderRepo      *repository.OrderRepository
	settlementRepo *repository.SettlementRepository
}

func NewSettlementService(orderRepo *repository.OrderRepository, settlementRepo *repository.SettlementRepository) *SettlementService {
	return &SettlementService{orderRepo: orderRepo, settlementRepo: settlementRepo}
}

type UnsettledSummary struct {
	Count          int                               `json:"count"`
	Total          float64                           `json:"total"`
	PaymentSummary map[string]PaymentMethodSummary   `json:"paymentSummary"`
}

type PaymentMethodSummary struct {
	Count int     `json:"count"`
	Total float64 `json:"total"`
}

func (s *SettlementService) GetUnsettledSummary(storeType, terminalID string) (*UnsettledSummary, error) {
	count, total, summary, err := s.orderRepo.FindUnsettledSummary(storeType, terminalID)
	if err != nil {
		return nil, err
	}
	paymentSummary := make(map[string]PaymentMethodSummary)
	for k, v := range summary {
		paymentSummary[k] = PaymentMethodSummary{Count: v.Count, Total: v.Total}
	}
	return &UnsettledSummary{
		Count:          count,
		Total:          total,
		PaymentSummary: paymentSummary,
	}, nil
}

func (s *SettlementService) Settle(settlementID string, timestamp int64, storeType, terminalID string) (*model.Settlement, error) {
	orders, err := s.orderRepo.FindUnsettledOrders(storeType, terminalID)
	if err != nil {
		return nil, err
	}
	if len(orders) == 0 {
		return nil, fmt.Errorf("NO_UNSETTLED_ORDERS")
	}

	var result *model.Settlement
	err = s.orderRepo.Transaction(func(tx *gorm.DB) error {
		txOrderRepo := s.orderRepo.WithTx(tx)
		txSettlementRepo := s.settlementRepo.WithTx(tx)

		var cardTotal, qrisTotal, cashTotal float64
		orderIDs := make([]string, 0, len(orders))
		for _, o := range orders {
			orderIDs = append(orderIDs, o.ID)
			switch o.PaymentMethod {
			case "CARD":
				cardTotal += o.Total
			case "QRIS":
				qrisTotal += o.Total
			case "CASH":
				cashTotal += o.Total
			}
		}

		totalAmount := cardTotal + qrisTotal + cashTotal
		settlement := &model.Settlement{
			ID:          settlementID,
			Timestamp:   timestamp,
			BatchCount:  len(orders),
			TotalAmount: totalAmount,
			CardTotal:   cardTotal,
			QRISTotal:   qrisTotal,
			CashTotal:   cashTotal,
			Status:      "SUCCESS",
			StoreType:   storeType,
			TerminalID:  terminalID,
			CreatedAt:   time.Now(),
		}
		if err := txSettlementRepo.Create(settlement); err != nil {
			return err
		}
		if err := txOrderRepo.MarkSettled(orderIDs); err != nil {
			return err
		}
		result = settlement
		return nil
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *SettlementService) List(limit int, storeType string) ([]model.Settlement, error) {
	return s.settlementRepo.FindAll(limit, storeType)
}

func (s *SettlementService) Get(id string) (*model.Settlement, error) {
	return s.settlementRepo.FindByID(id)
}
