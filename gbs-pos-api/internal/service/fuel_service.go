package service

import (
	"encoding/base64"
	"errors"
	"fmt"
	"gbs-pos-api/internal/dto"
	"gbs-pos-api/internal/model"
	"gbs-pos-api/internal/repository"
	gbsTemplate "gbs-pos-api/internal/template"
	"html/template"
	"strconv"
	"strings"
	"time"

	qrcode "github.com/skip2/go-qrcode"
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
func (s *FuelService) CreateSale(req dto.FuelSaleRequest) (*dto.FuelSaleResponse, bool, error) {
	// Idempotent: retries (e.g. lost response on the kiosk) must not create duplicates.
	if existing, err := s.saleRepo.FindByID(req.ID); err == nil {
		// Replay of an existing sale. Do NOT leak the authorization token (a bearer
		// capability for the dispenser) to a caller that only guessed the ID: it is
		// only returned when this request actually created the sale.
		resp := toFuelSaleResponse(existing)
		resp.AuthorizationToken = ""
		return resp, false, nil
	}

	var ts time.Time
	if req.Timestamp > 0 {
		ts = time.UnixMilli(req.Timestamp)
	} else {
		ts = time.Now()
	}

	receiptToken := req.ReceiptToken
	if receiptToken == "" {
		tok, err := GenerateToken()
		if err != nil {
			return nil, false, err
		}
		receiptToken = tok
	}
	authToken, err := GenerateToken()
	if err != nil {
		return nil, false, err
	}

	sale := &model.FuelSale{
		ID:                 req.ID,
		PumpID:             req.PumpID,
		NozzleID:           req.NozzleID,
		FuelCode:           req.FuelCode,
		PricePerLiter:      req.PricePerLiter,
		Liters:             req.Liters,
		TotalAmount:        req.TotalAmount,
		PaymentMethod:      req.PaymentMethod,
		TransactionID:      req.TransactionID,
		PosMessageID:       req.PosMessageID,
		Timestamp:          ts,
		ReceiptToken:       receiptToken,
		AuthorizationToken: authToken,
		Status:             model.FuelSaleStatusPaid,
	}
	if err := s.saleRepo.Create(sale); err != nil {
		// Lost a create race with a concurrent duplicate (same PK): fall back to the
		// idempotent replay path instead of failing with a duplicate-key 500.
		if existing, findErr := s.saleRepo.FindByID(req.ID); findErr == nil {
			resp := toFuelSaleResponse(existing)
			resp.AuthorizationToken = ""
			return resp, false, nil
		}
		return nil, false, err
	}
	return toFuelSaleResponse(sale), true, nil
}

func toFuelSaleResponse(sale *model.FuelSale) *dto.FuelSaleResponse {
	return &dto.FuelSaleResponse{
		ID:                 sale.ID,
		PumpID:             sale.PumpID,
		NozzleID:           sale.NozzleID,
		FuelCode:           sale.FuelCode,
		PricePerLiter:      sale.PricePerLiter,
		Liters:             sale.Liters,
		TotalAmount:        sale.TotalAmount,
		PaymentMethod:      sale.PaymentMethod,
		TransactionID:      sale.TransactionID,
		PosMessageID:       sale.PosMessageID,
		Timestamp:          sale.Timestamp,
		ReceiptToken:       sale.ReceiptToken,
		AuthorizationToken: sale.AuthorizationToken,
		Status:             sale.Status,
		AuthorizedAt:       sale.AuthorizedAt,
	}
}

func toPublicReceiptResponse(sale *model.FuelSale) dto.PublicReceiptResponse {
	return dto.PublicReceiptResponse{
		ID:            sale.ID,
		PumpID:        sale.PumpID,
		NozzleID:      sale.NozzleID,
		FuelCode:      sale.FuelCode,
		PricePerLiter: sale.PricePerLiter,
		Liters:        sale.Liters,
		TotalAmount:   sale.TotalAmount,
		PaymentMethod: sale.PaymentMethod,
		Timestamp:     sale.Timestamp,
		Status:        sale.Status,
		AuthorizedAt:  sale.AuthorizedAt,
	}
}

// GetPublicReceipt looks up a sale by its public receipt token.
func (s *FuelService) GetPublicReceipt(receiptToken string) (*dto.PublicReceiptResponse, error) {
	sale, err := s.saleRepo.FindByReceiptToken(receiptToken)
	if err != nil {
		return nil, err
	}
	res := toPublicReceiptResponse(sale)
	return &res, nil
}

// ValidateAuthorization checks a dispenser authorization token without side effects.
func (s *FuelService) ValidateAuthorization(authToken string) (*dto.AuthorizationCheckResponse, error) {
	sale, err := s.saleRepo.FindByAuthorizationToken(authToken)
	if err != nil {
		return nil, err
	}
	return &dto.AuthorizationCheckResponse{
		Valid:        true,
		Status:       sale.Status,
		SaleID:       sale.ID,
		PumpID:       sale.PumpID,
		NozzleID:     sale.NozzleID,
		FuelCode:     sale.FuelCode,
		Liters:       sale.Liters,
		TotalAmount:  sale.TotalAmount,
		AuthorizedAt: sale.AuthorizedAt,
	}, nil
}

// Authorize performs the single-use PAID → AUTHORIZED transition.
// Returns (response, alreadyUsed, error); alreadyUsed=true when the token was
// scanned before (status no longer PAID).
func (s *FuelService) Authorize(authToken string) (*dto.AuthorizationCheckResponse, bool, error) {
	sale, err := s.saleRepo.FindByAuthorizationToken(authToken)
	if err != nil {
		return nil, false, err
	}
	if sale.Status != model.FuelSaleStatusPaid {
		// Already authorized/used — report so the caller can distinguish a second scan.
		return &dto.AuthorizationCheckResponse{
			Valid:        false,
			Status:       sale.Status,
			SaleID:       sale.ID,
			PumpID:       sale.PumpID,
			NozzleID:     sale.NozzleID,
			FuelCode:     sale.FuelCode,
			Liters:       sale.Liters,
			TotalAmount:  sale.TotalAmount,
			AuthorizedAt: sale.AuthorizedAt,
		}, true, nil
	}

	now := time.Now()
	rows, err := s.saleRepo.Authorize(sale.ID, now)
	if err != nil {
		return nil, false, err
	}
	if rows == 0 {
		// Lost a race with another scanner.
		updated, findErr := s.saleRepo.FindByAuthorizationToken(authToken)
		if findErr != nil {
			return nil, true, nil
		}
		return &dto.AuthorizationCheckResponse{
			Valid:        false,
			Status:       updated.Status,
			SaleID:       updated.ID,
			PumpID:       updated.PumpID,
			NozzleID:     updated.NozzleID,
			FuelCode:     updated.FuelCode,
			Liters:       updated.Liters,
			TotalAmount:  updated.TotalAmount,
			AuthorizedAt: updated.AuthorizedAt,
		}, true, nil
	}

	return &dto.AuthorizationCheckResponse{
		Valid:        true,
		Status:       model.FuelSaleStatusAuthorized,
		SaleID:       sale.ID,
		PumpID:       sale.PumpID,
		NozzleID:     sale.NozzleID,
		FuelCode:     sale.FuelCode,
		Liters:       sale.Liters,
		TotalAmount:  sale.TotalAmount,
		AuthorizedAt: &now,
	}, false, nil
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

// RenderReceiptPage renders the public HTML receipt for a receipt token. It embeds a
// QR #2 (dispenser authorization) as a PNG data URI inside the page.
func (s *FuelService) RenderReceiptPage(receiptToken string) (string, error) {
	sale, err := s.saleRepo.FindByReceiptToken(receiptToken)
	if err != nil {
		return "", err
	}
	if sale.AuthorizationToken == "" {
		return "", errors.New("sale has no authorization token")
	}

	fuelName := sale.FuelCode
	if price, priceErr := s.priceRepo.FindByCode(sale.FuelCode); priceErr == nil {
		fuelName = price.Name
	}

	authQR, err := qrcode.Encode(sale.AuthorizationToken, qrcode.Medium, 512)
	if err != nil {
		return "", fmt.Errorf("generate authorization QR: %w", err)
	}
	authDataURI := template.URL("data:image/png;base64," + base64.StdEncoding.EncodeToString(authQR))

	wib := time.FixedZone("WIB", 7*60*60)
	alreadyUsed := sale.Status != model.FuelSaleStatusPaid
	authorizedAtStr := ""
	if sale.AuthorizedAt != nil {
		authorizedAtStr = sale.AuthorizedAt.In(wib).Format("02/01/2006 15:04")
	}

	data := gbsTemplate.ReceiptPageData{
		ID:            sale.ID,
		Timestamp:     sale.Timestamp.In(wib).Format("02/01/2006 15:04"),
		PumpID:        sale.PumpID,
		NozzleID:      sale.NozzleID,
		FuelName:      fuelName,
		PricePerLiter: formatRupiah(sale.PricePerLiter),
		Liters:        formatDecimal(sale.Liters) + " L",
		TotalAmount:   formatRupiah(sale.TotalAmount),
		PaymentMethod: displayPaymentMethod(sale.PaymentMethod),
		Status:        sale.Status,
		AuthorizedAt:  authorizedAtStr,
		AuthQRDataURI: authDataURI,
		AlreadyUsed:   alreadyUsed,
	}
	return gbsTemplate.RenderReceiptPage(data)
}

func formatRupiah(v float64) string {
	return "Rp " + formatDecimal(v)
}

func formatDecimal(v float64) string {
	s := strconv.FormatFloat(v, 'f', -1, 64)
	parts := strings.Split(s, ".")
	intPart := parts[0]
	var b strings.Builder
	n := len(intPart)
	for i, ch := range intPart {
		if i > 0 && (n-i)%3 == 0 {
			b.WriteByte('.')
		}
		b.WriteRune(ch)
	}
	if len(parts) == 2 {
		b.WriteString("," + parts[1])
	}
	return b.String()
}

func displayPaymentMethod(m string) string {
	switch strings.ToUpper(m) {
	case "CARD":
		return "Kartu"
	case "QRIS":
		return "QRIS"
	case "CASH":
		return "Tunai"
	default:
		return m
	}
}

func IsFuelNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
