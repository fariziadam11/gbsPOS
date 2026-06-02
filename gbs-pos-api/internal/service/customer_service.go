package service

import (
	"gbs-pos-api/internal/model"
	"gbs-pos-api/internal/repository"
)

type CustomerService struct {
	repo *repository.CustomerRepository
}

func NewCustomerService(repo *repository.CustomerRepository) *CustomerService {
	return &CustomerService{repo: repo}
}

func (s *CustomerService) List(query string) ([]model.Customer, error) {
	return s.repo.FindAll(query)
}

func (s *CustomerService) Get(id uint) (*model.Customer, error) {
	return s.repo.FindByID(id)
}

func (s *CustomerService) GetByPhone(phone string) (*model.Customer, error) {
	return s.repo.FindByPhone(phone)
}

func (s *CustomerService) Create(customer *model.Customer) error {
	return s.repo.Create(customer)
}

func (s *CustomerService) Update(id uint, updates *model.Customer) (*model.Customer, error) {
	customer, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if updates.Name != "" {
		customer.Name = updates.Name
	}
	if updates.Phone != "" {
		customer.Phone = updates.Phone
	}
	if updates.Email != "" {
		customer.Email = updates.Email
	}
	if updates.Address != "" {
		customer.Address = updates.Address
	}
	if err := s.repo.Update(customer); err != nil {
		return nil, err
	}
	return customer, nil
}

func (s *CustomerService) AddLoyaltyPoints(id uint, points int) error {
	return s.repo.AddLoyaltyPoints(id, points)
}

func (s *CustomerService) GetOrderHistory(id uint) ([]model.Order, error) {
	return s.repo.FindOrders(id)
}

func (s *CustomerService) GetCustomerWithHistory(id uint) (*model.Customer, []model.Order, error) {
	customer, err := s.repo.FindByID(id)
	if err != nil {
		return nil, nil, err
	}
	orders, err := s.repo.FindOrders(id)
	if err != nil {
		return nil, nil, err
	}
	return customer, orders, nil
}
