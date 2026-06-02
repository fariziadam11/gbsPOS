package repository

import (
	"gbs-pos-api/internal/model"

	"gorm.io/gorm"
)

type CustomerRepository struct {
	db *gorm.DB
}

func NewCustomerRepository(db *gorm.DB) *CustomerRepository {
	return &CustomerRepository{db: db}
}

func (r *CustomerRepository) FindAll(query string) ([]model.Customer, error) {
	var customers []model.Customer
	q := r.db.Order("created_at DESC")
	if query != "" {
		q = q.Where("name ILIKE ? OR phone ILIKE ?", "%"+query+"%", "%"+query+"%")
	}
	if err := q.Find(&customers).Error; err != nil {
		return nil, err
	}
	return customers, nil
}

func (r *CustomerRepository) FindByID(id uint) (*model.Customer, error) {
	var customer model.Customer
	if err := r.db.First(&customer, id).Error; err != nil {
		return nil, err
	}
	return &customer, nil
}

func (r *CustomerRepository) FindByPhone(phone string) (*model.Customer, error) {
	var customer model.Customer
	if err := r.db.Where("phone = ?", phone).First(&customer).Error; err != nil {
		return nil, err
	}
	return &customer, nil
}

func (r *CustomerRepository) Create(customer *model.Customer) error {
	return r.db.Create(customer).Error
}

func (r *CustomerRepository) Update(customer *model.Customer) error {
	return r.db.Save(customer).Error
}

func (r *CustomerRepository) AddLoyaltyPoints(id uint, points int) error {
	return r.db.Model(&model.Customer{}).Where("id = ?", id).Update("loyalty_points", gorm.Expr("loyalty_points + ?", points)).Error
}

func (r *CustomerRepository) FindOrders(customerID uint) ([]model.Order, error) {
	var orders []model.Order
	if err := r.db.Where("customer_id = ?", customerID).Order("timestamp DESC").Preload("Items").Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}
