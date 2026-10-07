package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"gbs-pos-api/internal/dto"
	"gbs-pos-api/internal/model"
	"gbs-pos-api/internal/repository"
	"math"
	"time"

	"gorm.io/gorm"
)

var ErrOrderAlreadyExists = errors.New("ORDER_ALREADY_EXISTS")
var ErrBTNQRISPaymentNotVerified = errors.New("BTN QRIS payment is not verified")

type OrderService struct {
	repo            *repository.OrderRepository
	productService  *ProductService
	customerService *CustomerService
	variantService  *ProductVariantService
	qrisTxRepo      *repository.QrisTransactionRepository
	btnQrisEnabled  bool
	posTaxRate      float64
}

func NewOrderService(
	repo *repository.OrderRepository,
	productService *ProductService,
	customerService *CustomerService,
	variantService *ProductVariantService,
) *OrderService {
	return &OrderService{
		repo:            repo,
		productService:  productService,
		customerService: customerService,
		variantService:  variantService,
	}
}

func (s *OrderService) ConfigureBTNQRISValidation(repo *repository.QrisTransactionRepository, enabled bool, taxRate float64) {
	s.qrisTxRepo = repo
	s.btnQrisEnabled = enabled
	s.posTaxRate = taxRate
}

func (s *OrderService) ValidateBTNCheckoutSnapshot(snapshot dto.QrisCheckoutSnapshotRequest) error {
	if !s.btnQrisEnabled || s.productService == nil {
		return fmt.Errorf("BTN checkout validator is not configured")
	}
	if err := validateQrisCheckoutSnapshot(snapshot); err != nil {
		return err
	}
	var retailSubtotal int64
	productQuantities := make(map[int]int)
	productStocks := make(map[int]int)
	variantQuantities := make(map[int]int)
	variantStocks := make(map[int]int)
	for _, item := range snapshot.Items {
		product, variant, err := s.productService.GetProductWithVariantDiscounts(item.ProductID, item.VariantID)
		if err != nil {
			return fmt.Errorf("checkout product is no longer available")
		}
		serverPrice := product.Price
		if item.VariantID != nil {
			if variant == nil || !variant.IsActive || variant.ProductID != item.ProductID {
				return fmt.Errorf("checkout variant is no longer available")
			}
			if variant.Price != nil {
				serverPrice = *variant.Price
			}
			variantQuantities[*item.VariantID] += item.Qty
			variantStocks[*item.VariantID] = variant.StockQuantity
		}
		clientPrice, err := rupiahCents(item.ProductPrice)
		if err != nil || clientPrice != cents(serverPrice) {
			return fmt.Errorf("checkout price no longer matches catalog")
		}
		productQuantities[item.ProductID] += item.Qty
		productStocks[item.ProductID] = product.StockQuantity
		line, _ := rupiahCents(item.Subtotal)
		retailSubtotal += line
	}
	for productID, quantity := range productQuantities {
		if productStocks[productID] < quantity {
			return fmt.Errorf("insufficient product stock")
		}
	}
	for variantID, quantity := range variantQuantities {
		if variantStocks[variantID] < quantity {
			return fmt.Errorf("insufficient variant stock")
		}
	}

	ppobSubtotal := int64(0)
	for _, item := range snapshot.PpobItems {
		ppobSubtotal += item.Total * 100
	}
	subtotal, _ := rupiahCents(snapshot.Subtotal)
	tax, _ := rupiahCents(snapshot.Tax)
	total, _ := rupiahCents(snapshot.Total)
	if subtotal != retailSubtotal+ppobSubtotal || tax != int64(math.Round(float64(retailSubtotal)*s.posTaxRate)) {
		return fmt.Errorf("checkout subtotal or tax does not match server calculation")
	}
	preDiscount := subtotal + tax
	discount := int64(0)
	if snapshot.DiscountType != "" && snapshot.DiscountValue != nil {
		value := *snapshot.DiscountValue
		switch snapshot.DiscountType {
		case "PERCENTAGE":
			if value < 0 || value > 100 {
				return fmt.Errorf("invalid checkout discount percentage")
			}
			discount = int64(math.Round(float64(preDiscount) * value / 100))
		case "FIXED":
			discount = cents(value)
		default:
			return fmt.Errorf("invalid checkout discount type")
		}
	}
	if snapshot.DiscountAmount != nil && cents(*snapshot.DiscountAmount) != discount {
		return fmt.Errorf("checkout discount amount does not match server calculation")
	}
	expectedTotal := preDiscount - discount
	if expectedTotal < 0 {
		expectedTotal = 0
	}
	if total != expectedTotal {
		return fmt.Errorf("checkout total does not match server calculation")
	}
	return nil
}

func cents(amount float64) int64 {
	return int64(math.Round(amount * 100))
}

func (s *OrderService) List(
	storeType string,
	startDate, endDate int64,
	isVoided, isSettled *bool,
	paymentMethod, terminalID string,
) ([]model.Order, error) {
	return s.repo.FindAll(
		storeType,
		startDate,
		endDate,
		isVoided,
		isSettled,
		paymentMethod,
		terminalID,
	)
}

func (s *OrderService) Get(id string) (*model.Order, error) {
	return s.repo.FindByIDWithItems(id)
}

func (s *OrderService) Create(order *model.Order) (*model.Order, bool, error) {
	existing, err := s.repo.FindByID(order.ID)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, false, err
		}
	} else if existing != nil {
		fullOrder, err := s.repo.FindByIDWithItems(order.ID)
		if err != nil {
			return nil, false, err
		}
		return fullOrder, true, nil
	}
	if err := s.validateBTNQRISPayment(order); err != nil {
		return nil, false, err
	}

	// Resolve customer by phone if no CustomerID provided
	if order.CustomerID == nil && order.CustomerPhone != "" {
		customer, err := s.customerService.GetByPhone(order.CustomerPhone)
		if err != nil {
			// Customer not found — create new one
			newCustomer := &model.Customer{
				Name:  order.CustomerName,
				Phone: order.CustomerPhone,
			}
			if newCustomer.Name == "" {
				newCustomer.Name = "Pelanggan " + order.CustomerPhone
			}
			if err := s.customerService.Create(newCustomer); err == nil {
				cid := int(newCustomer.ID)
				order.CustomerID = &cid
			}
		} else {
			cid := int(customer.ID)
			order.CustomerID = &cid
		}
	}

	// Calculate loyalty points: 1% of total (rounded down)
	loyaltyPoints := int(order.Total / 100)
	if loyaltyPoints < 1 {
		loyaltyPoints = 0
	}
	order.LoyaltyPointsEarned = loyaltyPoints

	// Calculate discount from the pre-discount subtotal and tax.
	if order.DiscountType != "" && order.DiscountValue != nil && *order.DiscountValue > 0 {
		preDiscountTotal := order.Subtotal + order.Tax
		var discountAmount float64
		if order.DiscountType == "PERCENTAGE" {
			discountAmount = preDiscountTotal * (*order.DiscountValue) / 100
		} else {
			discountAmount = *order.DiscountValue
		}
		if order.DiscountAmount != nil && math.Abs(*order.DiscountAmount-discountAmount) > 0.01 {
			return nil, false, fmt.Errorf("discount amount does not match discount configuration")
		}
		order.DiscountAmount = &discountAmount
		order.Total = preDiscountTotal - discountAmount
		if order.Total < 0 {
			order.Total = 0
		}
		// Recalculate loyalty points on discounted total
		loyaltyPoints = int(order.Total / 100)
		if loyaltyPoints < 1 {
			loyaltyPoints = 0
		}
		order.LoyaltyPointsEarned = loyaltyPoints
	}

	// Deduct stock in transaction
	if err := s.repo.Transaction(func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)

		// Lock the order row to prevent race condition on idempotency check
		// This ensures only one request can create an order with a given ID
		lockErr := tx.Raw("SELECT id FROM orders WHERE id = ? FOR UPDATE", order.ID).Error
		if lockErr != nil && !errors.Is(lockErr, gorm.ErrRecordNotFound) {
			return lockErr
		}

		// Check if order already exists after acquiring lock
		existingCheck, _ := txRepo.FindByID(order.ID)
		if existingCheck != nil {
			return ErrOrderAlreadyExists
		}

		if err := txRepo.Create(order); err != nil {
			return err
		}
		for _, item := range order.Items {
			if item.VariantID != nil {
				if err := s.variantService.DeductVariantStock(tx, *item.VariantID, item.Qty, order.ID); err != nil {
					return err
				}
			}
			if err := s.productService.DeductStock(tx, item.ProductID, item.Qty, order.ID); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		if errors.Is(err, ErrOrderAlreadyExists) {
			// Return idempotent response
			existingOrder, _ := s.repo.FindByIDWithItems(order.ID)
			if existingOrder != nil {
				return existingOrder, true, nil
			}
		}
		return nil, false, err
	}

	// Add loyalty points to customer (outside transaction to avoid lock contention)
	if order.CustomerID != nil && loyaltyPoints > 0 {
		if err := s.customerService.AddLoyaltyPoints(uint(*order.CustomerID), loyaltyPoints); err != nil {
			// Log error but don't fail the order creation
			// In production, use structured logging: log.Printf("failed to add loyalty points: %v", err)
		}
	}

	return order, false, nil
}

func (s *OrderService) validateBTNQRISPayment(order *model.Order) error {
	if !s.btnQrisEnabled || order.PaymentMethod != "QRIS" || order.QrisPaymentID != "" {
		return nil
	}
	if s.qrisTxRepo == nil || order.TransactionID == "" {
		return fmt.Errorf("%w: transaction reference is required", ErrBTNQRISPaymentNotVerified)
	}
	tx, err := s.qrisTxRepo.FindByID(context.Background(), order.TransactionID)
	if err != nil || tx.Provider != model.QrisProviderBTNSnap {
		return fmt.Errorf("%w: BTN transaction not found", ErrBTNQRISPaymentNotVerified)
	}
	if tx.Status != model.QrisTransactionStatusPaid || tx.OrderID != order.ID || tx.Currency != "IDR" {
		return fmt.Errorf("%w: transaction status or order does not match", ErrBTNQRISPaymentNotVerified)
	}
	if math.Round(tx.TotalAmount*100) != math.Round(order.Total*100) {
		return fmt.Errorf("%w: paid amount does not match order", ErrBTNQRISPaymentNotVerified)
	}
	if order.TerminalID != "" && order.TerminalID != tx.TerminalID {
		return fmt.Errorf("%w: terminal does not match payment", ErrBTNQRISPaymentNotVerified)
	}
	if len(tx.CheckoutSnapshot) == 0 {
		return fmt.Errorf("%w: checkout snapshot is missing", ErrBTNQRISPaymentNotVerified)
	}
	var expected dto.QrisCheckoutSnapshotRequest
	if err := json.Unmarshal(tx.CheckoutSnapshot, &expected); err != nil {
		return fmt.Errorf("%w: stored checkout snapshot is invalid", ErrBTNQRISPaymentNotVerified)
	}
	actual, err := checkoutSnapshotFromOrder(order)
	if err != nil {
		return fmt.Errorf("%w: order snapshot is invalid", ErrBTNQRISPaymentNotVerified)
	}
	expectedJSON, _ := json.Marshal(expected)
	actualJSON, _ := json.Marshal(actual)
	if string(expectedJSON) != string(actualJSON) {
		return fmt.Errorf("%w: order does not match paid checkout snapshot", ErrBTNQRISPaymentNotVerified)
	}
	order.TerminalID = tx.TerminalID
	if order.BankName == "" {
		order.BankName = "BTN"
	}
	return nil
}

func checkoutSnapshotFromOrder(order *model.Order) (dto.QrisCheckoutSnapshotRequest, error) {
	items := make([]dto.QrisCheckoutItem, len(order.Items))
	for i, item := range order.Items {
		items[i] = dto.QrisCheckoutItem{
			ProductID: item.ProductID, ProductName: item.ProductName, ProductPrice: item.ProductPrice,
			Qty: item.Qty, Subtotal: item.Subtotal, VariantID: item.VariantID,
			VariantName: item.VariantName, SKU: item.SKU,
		}
	}
	var ppobItems []dto.QrisCheckoutPpobItem
	if len(order.PpobItems) > 0 && string(order.PpobItems) != "null" {
		if err := json.Unmarshal(order.PpobItems, &ppobItems); err != nil {
			return dto.QrisCheckoutSnapshotRequest{}, err
		}
	}
	return dto.QrisCheckoutSnapshotRequest{
		ID: order.ID, Items: items, PpobItems: ppobItems,
		Subtotal: order.Subtotal, Tax: order.Tax, Total: order.Total,
		DiscountType: order.DiscountType, DiscountValue: order.DiscountValue,
		DiscountAmount: order.DiscountAmount,
	}, nil
}

func (s *OrderService) Void(id, reason, voidedBy string) (*model.Order, error) {
	order, err := s.repo.FindByIDWithItems(id)
	if err != nil {
		return nil, fmt.Errorf("ORDER_NOT_FOUND")
	}
	if order.IsVoided {
		return nil, fmt.Errorf("ORDER_ALREADY_VOIDED")
	}
	if order.IsSettled {
		return nil, fmt.Errorf("ORDER_ALREADY_SETTLED")
	}

	now := time.Now()
	order.IsVoided = true
	order.VoidReason = reason
	order.VoidedBy = voidedBy
	order.VoidedAt = &now

	// Restore stock in transaction
	if err := s.repo.Transaction(func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)
		if err := txRepo.UpdateVoid(order); err != nil {
			return err
		}
		for _, item := range order.Items {
			if item.VariantID != nil {
				if err := s.variantService.RestoreVariantStock(tx, *item.VariantID, item.Qty, order.ID); err != nil {
					return err
				}
			}
			if err := s.productService.RestoreStock(tx, item.ProductID, item.Qty, order.ID); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	// Deduct loyalty points earned from this order
	if order.CustomerID != nil && order.LoyaltyPointsEarned > 0 {
		if err := s.customerService.AddLoyaltyPoints(uint(*order.CustomerID), -order.LoyaltyPointsEarned); err != nil {
			// Log error but don't fail the void operation
			// In production, use structured logging
		}
	}

	// Reload with items for consistent response format
	return s.repo.FindByIDWithItems(id)
}

func (s *OrderService) BulkCreate(orders []model.Order) (*dto.BulkSyncResult, error) {
	result := &dto.BulkSyncResult{
		Orders: make([]model.Order, 0, len(orders)),
	}
	for i := range orders {
		createdOrder, idempotent, err := s.Create(&orders[i])
		if err != nil {
			result.Failed++
			continue
		}
		if idempotent {
			result.Existing++
			result.Idempotent = true
		} else {
			result.Created++
		}
		result.Orders = append(result.Orders, *createdOrder)
	}
	return result, nil
}
