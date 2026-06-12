package service

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"

	"gbs-pos-api/internal/model"
	"gbs-pos-api/internal/repository"
)

type PosHoldService struct {
	repo *repository.PosHoldRepository
}

func NewPosHoldService(repo *repository.PosHoldRepository) *PosHoldService {
	return &PosHoldService{repo: repo}
}

func (s *PosHoldService) Hold(storeType, terminalID string, payload any, total float64) (*model.PosHoldSession, error) {

	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	session := &model.PosHoldSession{
		ID:         uuid.NewString(),
		StoreType:  storeType,
		TerminalID: terminalID,
		Payload:    raw,
		Total:      total,
		Status:     model.PosHoldStatusActive,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := s.repo.Create(session); err != nil {
		return nil, err
	}

	return session, nil
}

func (s *PosHoldService) Get(id string) (*model.PosHoldSession, error) {
	return s.repo.FindByID(id)
}

func (s *PosHoldService) List(terminalID string) ([]model.PosHoldSession, error) {
	return s.repo.FindActiveByTerminal(terminalID)
}

func (s *PosHoldService) Resume(id string) (*model.PosHoldSession, error) {

	session, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if session.Status != model.PosHoldStatusActive {
		return nil, errors.New("CANNOT_RESUME_NON_ACTIVE_HOLD")
	}

	session.Status = model.PosHoldStatusResumed
	session.UpdatedAt = time.Now()

	if err := s.repo.UpdateStatus(session.ID, session.Status); err != nil {
		return nil, err
	}

	return session, nil
}

func (s *PosHoldService) Delete(id string) error {
	return s.repo.Delete(id)
}