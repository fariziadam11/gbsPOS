package service

import (
	"encoding/json"
	"fmt"
	"gbs-pos-api/internal/dto"
	"gbs-pos-api/internal/model"
)

func ValidateOrder(order *model.Order) error {
	if order.ID == "" {
		return fmt.Errorf("VALIDATION_ERROR: id is required")
	}
	if order.Subtotal < 0 {
		return fmt.Errorf("VALIDATION_ERROR: subtotal must be >= 0")
	}
	if order.Tax < 0 {
		return fmt.Errorf("VALIDATION_ERROR: tax must be >= 0")
	}
	if order.Total < 0 {
		return fmt.Errorf("VALIDATION_ERROR: total must be >= 0")
	}
	validPaymentMethods := map[string]bool{"CASH": true, "CARD": true, "QRIS": true}
	if !validPaymentMethods[order.PaymentMethod] {
		return fmt.Errorf("VALIDATION_ERROR: paymentMethod must be one of: CASH, CARD, QRIS")
	}
	validStoreTypes := map[string]bool{"RETAIL": true, "FNB": true, "OUTFIT": true}
	if order.StoreType != "" && !validStoreTypes[order.StoreType] {
		return fmt.Errorf("VALIDATION_ERROR: storeType must be one of: RETAIL, FNB, OUTFIT")
	}
	var ppobItems []dto.QrisCheckoutPpobItem
	if len(order.PpobItems) > 0 && string(order.PpobItems) != "null" {
		if err := json.Unmarshal(order.PpobItems, &ppobItems); err != nil {
			return fmt.Errorf("VALIDATION_ERROR: invalid ppobItems")
		}
	}
	if len(order.Items) == 0 && len(ppobItems) == 0 {
		return fmt.Errorf("VALIDATION_ERROR: items cannot be empty")
	}
	for _, item := range order.Items {
		if item.Qty <= 0 {
			return fmt.Errorf("VALIDATION_ERROR: item qty must be > 0")
		}
		if item.ProductName == "" {
			return fmt.Errorf("VALIDATION_ERROR: item productName is required")
		}
	}
	for _, item := range ppobItems {
		if item.PpobProductID == "" || item.CustomerID == "" || item.Amount < 0 || item.AdminFee < 0 || item.Total != item.Amount+item.AdminFee {
			return fmt.Errorf("VALIDATION_ERROR: invalid PPOB item")
		}
	}
	return nil
}
