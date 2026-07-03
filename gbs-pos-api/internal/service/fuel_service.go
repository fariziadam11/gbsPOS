package service

import (
	"errors"
	"gbs-pos-api/internal/dto"
	"gbs-pos-api/internal/model"
	"gbs-pos-api/internal/repository"
	"time"

	"gorm.io/gorm"
)

type FuelService struct {
	priceRepo *repository.FuelPriceRepository
	pumpRepo  *repository.PumpRepository
	nozzleRepo *repository.NozzleRepository
	saleRepo  *repository.FuelSaleRepository
}

func NewFuelService(
	priceRepo *repository.FuelPriceRepository,
	pumpRepo *repository.PumpRepository,
	nozzleRepo *repository.NozzleRepository,
	saleRepo *repository.FuelSaleRepository,
) *FuelService {
	return &FuelService{
		priceRepo: priceRepo,
		pumpRepo:  pumpRepo,
		nozzleRepo: nozzleRepo,
		saleRepo:  saleRepo,
	}
}

// Fuel prices
func (s *FuelService) ListPrices() ([]dto.FuelPriceResponse, error) {
	prices, err := s.priceRepo.FindAll()
	if err != nil {
		return nil, err
	}
	var res []dto.FuelPriceResponse
	for _, p := range prices {
		res = append(res, dto.FuelPriceResponse{
			Code:          p.Code,
			Name:          p.Name,
			PricePerLiter: p.PricePerLiter,
			UpdatedAt:     p.UpdatedAt.UnixMilli(),
		})
	}
	return res, nil
}

func (s *FuelService) UpdatePrice(code string, req dto.UpdateFuelPriceRequest) (*dto.FuelPriceResponse, error) {
	price, err := s.priceRepo.FindByCode(code)
	if err != nil {
		return nil, err
	}
	price.PricePerLiter = req.PricePerLiter
	price.UpdatedAt = time.Now()
	if err := s.priceRepo.Update(price); err != nil {
		return nil, err
	}
	return &dto.FuelPriceResponse{
		Code:          price.Code,
		Name:          price.Name,
		PricePerLiter: price.PricePerLiter,
		UpdatedAt:     price.UpdatedAt.UnixMilli(),
	}, nil
}

// Pumps
func (s *FuelService) ListPumps() ([]dto.PumpResponse, error) {
	pumps, err := s.pumpRepo.FindAll()
	if err != nil {
		return nil, err
	}
	var res []dto.PumpResponse
	for _, p := range pumps {
		res = append(res, dto.PumpResponse{ID: p.ID, Name: p.Name, IsActive: p.IsActive})
	}
	return res, nil
}

func (s *FuelService) CreatePump(req dto.CreatePumpRequest) (*dto.PumpResponse, error) {
	pump := &model.Pump{ID: req.ID, Name: req.Name, IsActive: true}
	if err := s.pumpRepo.Create(pump); err != nil {
		return nil, err
	}
	return &dto.PumpResponse{ID: pump.ID, Name: pump.Name, IsActive: pump.IsActive}, nil
}

func (s *FuelService) UpdatePump(id string, req dto.UpdatePumpRequest) (*dto.PumpResponse, error) {
	pump, err := s.pumpRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if req.Name != "" {
		pump.Name = req.Name
	}
	if req.IsActive != nil {
		pump.IsActive = *req.IsActive
	}
	if err := s.pumpRepo.Update(pump); err != nil {
		return nil, err
	}
	return &dto.PumpResponse{ID: pump.ID, Name: pump.Name, IsActive: pump.IsActive}, nil
}

func (s *FuelService) DeletePump(id string) error {
	return s.pumpRepo.Delete(id)
}

// Nozzles
func (s *FuelService) ListNozzles() ([]dto.NozzleResponse, error) {
	nozzles, err := s.nozzleRepo.FindAll()
	if err != nil {
		return nil, err
	}
	var res []dto.NozzleResponse
	for _, n := range nozzles {
		res = append(res, dto.NozzleResponse{
			ID:       n.ID,
			PumpID:   n.PumpID,
			Name:     n.Name,
			FuelCode: n.FuelCode,
			IsActive: n.IsActive,
		})
	}
	return res, nil
}

func (s *FuelService) CreateNozzle(req dto.CreateNozzleRequest) (*dto.NozzleResponse, error) {
	nozzle := &model.Nozzle{
		ID:       req.ID,
		PumpID:   req.PumpID,
		Name:     req.Name,
		FuelCode: req.FuelCode,
		IsActive: true,
	}
	if err := s.nozzleRepo.Create(nozzle); err != nil {
		return nil, err
	}
	return &dto.NozzleResponse{
		ID:       nozzle.ID,
		PumpID:   nozzle.PumpID,
		Name:     nozzle.Name,
		FuelCode: nozzle.FuelCode,
		IsActive: nozzle.IsActive,
	}, nil
}

func (s *FuelService) UpdateNozzle(id string, req dto.UpdateNozzleRequest) (*dto.NozzleResponse, error) {
	nozzle, err := s.nozzleRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if req.Name != "" {
		nozzle.Name = req.Name
	}
	if req.FuelCode != "" {
		nozzle.FuelCode = req.FuelCode
	}
	if req.IsActive != nil {
		nozzle.IsActive = *req.IsActive
	}
	if err := s.nozzleRepo.Update(nozzle); err != nil {
		return nil, err
	}
	return &dto.NozzleResponse{
		ID:       nozzle.ID,
		PumpID:   nozzle.PumpID,
		Name:     nozzle.Name,
		FuelCode: nozzle.FuelCode,
		IsActive: nozzle.IsActive,
	}, nil
}

func (s *FuelService) DeleteNozzle(id string) error {
	return s.nozzleRepo.Delete(id)
}

// Fuel sales
func (s *FuelService) CreateSale(req dto.FuelSaleRequest) (*dto.FuelSaleResponse, error) {
	var ts time.Time
	if req.Timestamp > 0 {
		ts = time.UnixMilli(req.Timestamp)
	} else {
		ts = time.Now()
	}
	sale := &model.FuelSale{
		ID:            req.ID,
		PumpID:        req.PumpID,
		NozzleID:      req.NozzleID,
		FuelCode:      req.FuelCode,
		PricePerLiter: req.PricePerLiter,
		Liters:        req.Liters,
		TotalAmount:   req.TotalAmount,
		PaymentMethod: req.PaymentMethod,
		TransactionID: req.TransactionID,
		PosMessageID:  req.PosMessageID,
		Timestamp:     ts,
	}
	if err := s.saleRepo.Create(sale); err != nil {
		return nil, err
	}
	return &dto.FuelSaleResponse{
		ID:            sale.ID,
		PumpID:        sale.PumpID,
		NozzleID:      sale.NozzleID,
		FuelCode:      sale.FuelCode,
		PricePerLiter: sale.PricePerLiter,
		Liters:        sale.Liters,
		TotalAmount:   sale.TotalAmount,
		PaymentMethod: sale.PaymentMethod,
		TransactionID: sale.TransactionID,
		PosMessageID:  sale.PosMessageID,
		Timestamp:     sale.Timestamp,
	}, nil
}

func (s *FuelService) Report(from, to time.Time) (*dto.FuelSalesReportResponse, error) {
	report, err := s.saleRepo.Report(from, to)
	if err != nil {
		return nil, err
	}
	res := &dto.FuelSalesReportResponse{}
	for _, item := range report.Summary {
		res.Summary = append(res.Summary, dto.FuelReportItem{
			FuelCode:    item.FuelCode,
			Liters:      item.Liters,
			TotalAmount: item.TotalAmount,
		})
	}
	for _, item := range report.PumpTotals {
		res.PumpTotals = append(res.PumpTotals, dto.PumpReportItem{
			PumpID:      item.PumpID,
			Liters:      item.Liters,
			TotalAmount: item.TotalAmount,
		})
	}
	return res, nil
}

func IsFuelNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
