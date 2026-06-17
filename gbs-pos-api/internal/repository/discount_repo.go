package repository

import (
	"gbs-pos-api/internal/model"
	"time"

	"gorm.io/gorm"
)

type DiscountRepository struct {
	db *gorm.DB
}

func NewDiscountRepository(db *gorm.DB) *DiscountRepository {
	return &DiscountRepository{db: db}
}

func (r *DiscountRepository) Create(discount *model.Discount) error {
	return r.db.Create(discount).Error
}

func (r *DiscountRepository) Update(discount *model.Discount) error {
	return r.db.Save(discount).Error
}

func (r *DiscountRepository) FindByID(id uint) (*model.Discount, error) {
	var discount model.Discount
	if err := r.db.First(&discount, id).Error; err != nil {
		return nil, err
	}
	return &discount, nil
}

func (r *DiscountRepository) FindByProductID(productID uint) ([]model.Discount, error) {
	var discounts []model.Discount
	if err := r.db.
		Where("scope = ?", model.DiscountScopeProduct).
		Where("product_id = ?", productID).
		Order("start_date DESC").
		Find(&discounts).Error; err != nil {
		return nil, err
	}
	return discounts, nil
}

func (r *DiscountRepository) FindAll() ([]model.Discount, error) {
	var discounts []model.Discount
	if err := r.db.Order("created_at DESC").Find(&discounts).Error; err != nil {
		return nil, err
	}
	return discounts, nil
}

func (r *DiscountRepository) Delete(id uint) error {
	return r.db.Delete(&model.Discount{}, id).Error
}

func (r *DiscountRepository) FindOverlappingDiscount(
	productID uint,
	excludeID uint,
	startDate time.Time,
	endDate time.Time,
	now time.Time,
) (*model.Discount, error) {
	var discount model.Discount
	query := r.db.
		Where("scope = ?", model.DiscountScopeProduct).
		Where("product_id = ?", productID).
		Where("status NOT IN ?", []string{"STOPPED", "CANCELLED"}).
		Where("end_date >= ?", now).
		Where("start_date <= ? AND end_date >= ?", endDate, startDate).
		Order("start_date ASC")
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}
	if err := query.First(&discount).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &discount, nil
}

func (r *DiscountRepository) FindActiveByProductIDs(
	productIDs []uint,
	now time.Time,
) ([]model.Discount, error) {
	if len(productIDs) == 0 {
		return []model.Discount{}, nil
	}

	var discounts []model.Discount
	if err := r.db.
		Where("scope = ?", model.DiscountScopeProduct).
		Where("product_id IN ?", productIDs).
		Where("status NOT IN ?", []string{"STOPPED", "CANCELLED"}).
		Where("start_date <= ? AND end_date >= ?", now, now).
		Order("product_id ASC, start_date DESC").
		Find(&discounts).Error; err != nil {
		return nil, err
	}
	return discounts, nil
}

func (r *DiscountRepository) FindActiveByScope(
	scope string,
	now time.Time,
) ([]model.Discount, error) {
	var discounts []model.Discount
	if err := r.db.
		Where("scope = ?", scope).
		Where("status NOT IN ?", []string{"STOPPED", "CANCELLED"}).
		Where("start_date <= ? AND end_date >= ?", now, now).
		Order("start_date DESC").
		Find(&discounts).Error; err != nil {
		return nil, err
	}
	return discounts, nil
}

func (r *DiscountRepository) FindTransactionDiscount() ([]model.Discount, error) {
	var discounts []model.Discount
	if err := r.db.
		Where("scope = ?", model.DiscountScopeTransaction).
		Order("start_date DESC").
		Find(&discounts).Error; err != nil {
		return nil, err
	}
	return discounts, nil
}

func (r *DiscountRepository) FindVoucherDiscountsByCode(code string) ([]model.Discount, error) {
	var discounts []model.Discount
	if err := r.db.
		Where("scope = ?", model.DiscountScopeVoucher).
		Where("voucher_code = ?", code).
		Order("start_date DESC").
		Find(&discounts).Error; err != nil {
		return nil, err
	}
	return discounts, nil
}

func (r *DiscountRepository) FindActiveVoucherDiscountsByCode(
	code string,
	now time.Time,
) ([]model.Discount, error) {
	var discounts []model.Discount
	if err := r.db.
		Where("scope = ?", model.DiscountScopeVoucher).
		Where("voucher_code = ?", code).
		Where("status NOT IN ?", []string{"STOPPED", "CANCELLED"}).
		Where("start_date <= ? AND end_date >= ?", now, now).
		Order("start_date DESC").
		Find(&discounts).Error; err != nil {
		return nil, err
	}
	return discounts, nil
}

func (r *DiscountRepository) FindVoucherByCode(code string) (*model.Discount, error) {
	var discount model.Discount
	if err := r.db.
		Where("scope = ?", model.DiscountScopeVoucher).
		Where("voucher_code = ?", code).
		Order("start_date DESC").
		First(&discount).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &discount, nil
}
