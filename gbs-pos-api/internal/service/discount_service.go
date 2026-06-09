package service

import (
	"errors"
	"fmt"
	"gbs-pos-api/internal/dto"
	"gbs-pos-api/internal/model"
	"gbs-pos-api/internal/repository"
	"strings"
	"time"

	"gorm.io/gorm"
)

const (
	DiscountTypePercentage = "PERCENTAGE"
	DiscountTypeFixed      = "FIXED"

	DiscountStatusPending   = "PENDING"
	DiscountStatusActive    = "ACTIVE"
	DiscountStatusExpired   = "EXPIRED"
	DiscountStatusStopped   = "STOPPED"
	DiscountStatusCancelled = "CANCELLED"
)

var (
	ErrDiscountNotFound        = errors.New("DISCOUNT_NOT_FOUND")
	ErrDiscountOverlap         = errors.New("DISCOUNT_PERIOD_OVERLAP")
	ErrDiscountInvalidStatus   = errors.New("INVALID_DISCOUNT_STATUS")
	ErrDiscountProductNotFound = errors.New("PRODUCT_NOT_FOUND")
)

type DiscountValidationError struct {
	Message string
}

func (e *DiscountValidationError) Error() string {
	return e.Message
}

type DiscountService struct {
	repo        *repository.DiscountRepository
	productRepo *repository.ProductRepository
	now         func() time.Time
}

func NewDiscountService(
	repo *repository.DiscountRepository,
	productRepo *repository.ProductRepository,
) *DiscountService {
	return &DiscountService{
		repo:        repo,
		productRepo: productRepo,
		now:         time.Now,
	}
}

func (s *DiscountService) Create(req dto.CreateDiscountRequest) (*dto.DiscountResponse, error) {
	startDate, endDate, err := parseDiscountDateRange(req.StartDate, req.EndDate)
	if err != nil {
		return nil, err
	}

	discount := &model.Discount{
		ProductID: req.ProductID,
		Name:      strings.TrimSpace(req.Name),
		Type:      strings.ToUpper(strings.TrimSpace(req.Type)),
		Value:     req.Value,
		StartDate: startDate,
		EndDate:   endDate,
	}

	if err := s.ensureProductExists(discount.ProductID); err != nil {
		return nil, err
	}
	if err := validateDiscount(discount); err != nil {
		return nil, err
	}

	discount.Status = s.calculateEffectiveStatus(discount, s.now())
	if err := s.ensureNoOverlap(discount, 0); err != nil {
		return nil, err
	}
	if err := s.repo.Create(discount); err != nil {
		return nil, err
	}
	return s.toResponse(discount), nil
}

func (s *DiscountService) Update(
	id uint,
	req dto.UpdateDiscountRequest,
) (*dto.DiscountResponse, error) {
	discount, err := s.findByID(id)
	if err != nil {
		return nil, err
	}

	if req.ProductID != nil {
		if *req.ProductID == 0 {
			return nil, &DiscountValidationError{Message: "productId is required"}
		}
		if err := s.ensureProductExists(*req.ProductID); err != nil {
			return nil, err
		}
		discount.ProductID = *req.ProductID
	}
	if strings.TrimSpace(req.Name) != "" {
		discount.Name = strings.TrimSpace(req.Name)
	}
	if strings.TrimSpace(req.Type) != "" {
		discount.Type = strings.ToUpper(strings.TrimSpace(req.Type))
	}
	if req.Value != nil {
		discount.Value = *req.Value
	}
	if strings.TrimSpace(req.StartDate) != "" {
		startDate, err := parseDiscountDate(req.StartDate, false)
		if err != nil {
			return nil, err
		}
		discount.StartDate = startDate
	}
	if strings.TrimSpace(req.EndDate) != "" {
		endDate, err := parseDiscountDate(req.EndDate, true)
		if err != nil {
			return nil, err
		}
		discount.EndDate = endDate
	}

	if err := validateDiscount(discount); err != nil {
		return nil, err
	}
	if !isTerminalDiscountStatus(discount.Status) {
		discount.Status = s.calculateEffectiveStatus(discount, s.now())
	}
	if err := s.ensureNoOverlap(discount, discount.ID); err != nil {
		return nil, err
	}
	if err := s.repo.Update(discount); err != nil {
		return nil, err
	}
	return s.toResponse(discount), nil
}

func (s *DiscountService) Get(id uint) (*dto.DiscountResponse, error) {
	discount, err := s.findByID(id)
	if err != nil {
		return nil, err
	}
	return s.toResponse(discount), nil
}

func (s *DiscountService) List(productID uint) ([]dto.DiscountResponse, error) {
	var (
		discounts []model.Discount
		err       error
	)
	if productID > 0 {
		discounts, err = s.repo.FindByProductID(productID)
	} else {
		discounts, err = s.repo.FindAll()
	}
	if err != nil {
		return nil, err
	}

	responses := make([]dto.DiscountResponse, 0, len(discounts))
	for i := range discounts {
		responses = append(responses, *s.toResponse(&discounts[i]))
	}
	return responses, nil
}

func (s *DiscountService) Delete(id uint) error {
	if _, err := s.findByID(id); err != nil {
		return err
	}
	return s.repo.Delete(id)
}

func (s *DiscountService) Stop(id uint) (*dto.DiscountResponse, error) {
	discount, err := s.findByID(id)
	if err != nil {
		return nil, err
	}
	if s.GetEffectiveStatus(discount) != DiscountStatusActive {
		return nil, ErrDiscountInvalidStatus
	}

	discount.Status = DiscountStatusStopped
	if err := s.repo.Update(discount); err != nil {
		return nil, err
	}
	return s.toResponse(discount), nil
}

func (s *DiscountService) Cancel(id uint) (*dto.DiscountResponse, error) {
	discount, err := s.findByID(id)
	if err != nil {
		return nil, err
	}
	if s.GetEffectiveStatus(discount) != DiscountStatusPending {
		return nil, ErrDiscountInvalidStatus
	}

	discount.Status = DiscountStatusCancelled
	if err := s.repo.Update(discount); err != nil {
		return nil, err
	}
	return s.toResponse(discount), nil
}

func (s *DiscountService) GetEffectiveStatus(discount *model.Discount) string {
	return s.calculateEffectiveStatus(discount, s.now())
}

func (s *DiscountService) GetActiveDiscountsByProductIDs(
	productIDs []uint,
) (map[uint]*model.Discount, error) {
	now := s.now()
	discounts, err := s.repo.FindActiveByProductIDs(productIDs, now)
	if err != nil {
		return nil, err
	}

	result := make(map[uint]*model.Discount)
	for i := range discounts {
		discount := &discounts[i]
		if s.calculateEffectiveStatus(discount, now) != DiscountStatusActive {
			continue
		}
		if _, exists := result[discount.ProductID]; exists {
			continue
		}
		result[discount.ProductID] = discount
	}
	return result, nil
}

func (s *DiscountService) ApplyDiscount(price float64, discount *model.Discount) float64 {
	if discount == nil || s.GetEffectiveStatus(discount) != DiscountStatusActive {
		return price
	}

	finalPrice := price
	switch discount.Type {
	case DiscountTypePercentage:
		finalPrice = price - (price * discount.Value / 100)
	case DiscountTypeFixed:
		finalPrice = price - discount.Value
	}
	if finalPrice < 0 {
		return 0
	}
	return finalPrice
}

func (s *DiscountService) findByID(id uint) (*model.Discount, error) {
	discount, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDiscountNotFound
		}
		return nil, err
	}
	return discount, nil
}

func (s *DiscountService) ensureProductExists(productID uint) error {
	if productID == 0 {
		return &DiscountValidationError{Message: "productId is required"}
	}
	if _, err := s.productRepo.FindByID(productID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrDiscountProductNotFound
		}
		return err
	}
	return nil
}

func (s *DiscountService) ensureNoOverlap(discount *model.Discount, excludeID uint) error {
	effectiveStatus := s.GetEffectiveStatus(discount)
	if effectiveStatus != DiscountStatusActive && effectiveStatus != DiscountStatusPending {
		return nil
	}

	overlapping, err := s.repo.FindOverlappingDiscount(
		discount.ProductID,
		excludeID,
		discount.StartDate,
		discount.EndDate,
		s.now(),
	)
	if err != nil {
		return err
	}
	if overlapping == nil {
		return nil
	}
	overlappingStatus := s.GetEffectiveStatus(overlapping)
	if overlappingStatus == DiscountStatusActive || overlappingStatus == DiscountStatusPending {
		return ErrDiscountOverlap
	}
	return nil
}

func (s *DiscountService) calculateEffectiveStatus(discount *model.Discount, now time.Time) string {
	if discount == nil {
		return ""
	}
	if isTerminalDiscountStatus(discount.Status) {
		return discount.Status
	}
	if now.Before(discount.StartDate) {
		return DiscountStatusPending
	}
	if now.After(discount.EndDate) {
		return DiscountStatusExpired
	}
	return DiscountStatusActive
}

func (s *DiscountService) toResponse(discount *model.Discount) *dto.DiscountResponse {
	return &dto.DiscountResponse{
		ID:              discount.ID,
		ProductID:       discount.ProductID,
		Name:            discount.Name,
		Type:            discount.Type,
		Value:           discount.Value,
		StartDate:       discount.StartDate,
		EndDate:         discount.EndDate,
		Status:          discount.Status,
		EffectiveStatus: s.GetEffectiveStatus(discount),
		CreatedAt:       discount.CreatedAt,
		UpdatedAt:       discount.UpdatedAt,
	}
}

func validateDiscount(discount *model.Discount) error {
	if discount.ProductID == 0 {
		return &DiscountValidationError{Message: "productId is required"}
	}
	if strings.TrimSpace(discount.Name) == "" {
		return &DiscountValidationError{Message: "name is required"}
	}

	switch discount.Type {
	case DiscountTypePercentage:
		if discount.Value <= 0 || discount.Value > 100 {
			return &DiscountValidationError{
				Message: "percentage discount value must be > 0 and <= 100",
			}
		}
	case DiscountTypeFixed:
		if discount.Value <= 0 {
			return &DiscountValidationError{Message: "fixed discount value must be > 0"}
		}
	default:
		return &DiscountValidationError{
			Message: "type must be one of: PERCENTAGE, FIXED",
		}
	}

	if discount.StartDate.IsZero() {
		return &DiscountValidationError{Message: "startDate is required"}
	}
	if discount.EndDate.IsZero() {
		return &DiscountValidationError{Message: "endDate is required"}
	}
	if discount.StartDate.After(discount.EndDate) {
		return &DiscountValidationError{Message: "startDate must be before or equal to endDate"}
	}
	return nil
}

func parseDiscountDateRange(startDate, endDate string) (time.Time, time.Time, error) {
	start, err := parseDiscountDate(startDate, false)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	end, err := parseDiscountDate(endDate, true)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	return start, end, nil
}

func parseDiscountDate(value string, endOfDay bool) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, &DiscountValidationError{Message: "date is required"}
	}

	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return parsed, nil
	}
	if parsed, err := time.ParseInLocation("2006-01-02 15:04:05", value, time.Local); err == nil {
		return parsed, nil
	}
	if parsed, err := time.ParseInLocation("2006-01-02", value, time.Local); err == nil {
		if endOfDay {
			return parsed.Add(24*time.Hour - time.Nanosecond), nil
		}
		return parsed, nil
	}

	return time.Time{}, &DiscountValidationError{
		Message: fmt.Sprintf("invalid date format %q, use RFC3339 or YYYY-MM-DD", value),
	}
}

func isTerminalDiscountStatus(status string) bool {
	return status == DiscountStatusStopped || status == DiscountStatusCancelled
}
