package service

import (
	"errors"

	"github.com/google/uuid"

	"gbs-pos-api/internal/dto"
	"gbs-pos-api/internal/model"
	"gbs-pos-api/internal/repository"

	"gorm.io/datatypes"
)

var (
	ErrHoldCannotResumeNonActive = errors.New("CANNOT_RESUME_NON_ACTIVE_HOLD")
	ErrHoldCannotDeleteNonActive = errors.New("CANNOT_DELETE_NON_ACTIVE_HOLD")
)

type HoldService struct {
	repo *repository.HoldRepository
}

func NewHoldService(repo *repository.HoldRepository) *HoldService {
	return &HoldService{repo: repo}
}

func (s *HoldService) Create(req dto.CreateHoldRequest) (*model.HoldSession, error) {
	session := &model.HoldSession{
		ID:         uuid.NewString(),
		StoreType:  req.StoreType,
		TerminalID: req.TerminalID,
		Payload:    datatypes.JSON(req.Payload),
		Total:      req.Total,
		Status:     model.HoldStatusActive,
	}

	if err := s.repo.Create(session); err != nil {
		return nil, err
	}

	return session, nil
}

func (s *HoldService) Get(id string) (*model.HoldSession, error) {
	return s.repo.FindByID(id)
}

func (s *HoldService) List(terminalID string) ([]model.HoldSession, error) {
	return s.repo.FindActiveByTerminal(terminalID)
}

func (s *HoldService) Resume(id string) (*model.HoldSession, error) {
	session, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if session.Status != model.HoldStatusActive {
		return nil, ErrHoldCannotResumeNonActive
	}

	if err := s.repo.UpdateStatus(session.ID, model.HoldStatusResumed); err != nil {
		return nil, err
	}

	return s.repo.FindByID(id)
}

func (s *HoldService) Delete(id string) error {
	session, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}
	if session.Status != model.HoldStatusActive {
		return ErrHoldCannotDeleteNonActive
	}
	return s.repo.Delete(id)
}
