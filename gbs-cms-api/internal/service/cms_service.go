package service

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"gbs-cms-api/internal/model"
	"gbs-cms-api/internal/repository"
)

type CMSService struct {
	adRepo      *repository.AdRepository
	playLogRepo *repository.AdPlayLogRepository
	uploadDir   string
}

func NewCMSService(adRepo *repository.AdRepository, playLogRepo *repository.AdPlayLogRepository, uploadDir string) *CMSService {
	return &CMSService{adRepo: adRepo, playLogRepo: playLogRepo, uploadDir: uploadDir}
}

type AdListResult struct {
	Ads        []model.Ad `json:"ads"`
	Pagination Pagination `json:"pagination"`
}

type Pagination struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"totalPages"`
}

func (s *CMSService) ListAds(page, limit int) (*AdListResult, error) {
	ads, total, err := s.adRepo.FindAll(page, limit)
	if err != nil {
		return nil, err
	}
	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}
	return &AdListResult{
		Ads: ads,
		Pagination: Pagination{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}

func (s *CMSService) GetAd(id uint) (*model.Ad, error) {
	return s.adRepo.FindByID(id)
}

func validateSchedule(startDate, endDate *time.Time) error {
	if startDate != nil && endDate != nil && startDate.After(*endDate) {
		return fmt.Errorf("INVALID_SCHEDULE")
	}
	return nil
}

func (s *CMSService) CreateAd(name, filename, mimeType string, fileSize int64, storeTypes []string, playlistOrder int, startDate, endDate, startTime, endTime *time.Time, createdBy uint) (*model.Ad, error) {
	baseName := strings.TrimSuffix(filename, filepath.Ext(filename))
	newFilename := fmt.Sprintf("%s_%d%s", baseName, time.Now().UnixMilli(), filepath.Ext(filename))
	storagePath := filepath.Join(s.uploadDir, newFilename)
	if err := validateSchedule(startDate, endDate); err != nil {
		return nil, err
	}
	ad := &model.Ad{
		Name:          name,
		Filename:      filename,
		StoragePath:   storagePath,
		FileSize:      fileSize,
		MimeType:      mimeType,
		StoreTypes:    storeTypes,
		PlaylistOrder: playlistOrder,
		IsActive:      true,
		StartDate:     startDate,
		EndDate:       endDate,
		StartTime:     startTime,
		EndTime:       endTime,
		CreatedBy:     createdBy,
	}
	if err := s.adRepo.Create(ad); err != nil {
		return nil, err
	}
	return ad, nil
}

func (s *CMSService) SaveUpload(src io.Reader, storagePath string) error {
	if err := os.MkdirAll(filepath.Dir(storagePath), 0755); err != nil {
		return err
	}
	out, err := os.Create(storagePath)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, src)
	return err
}

func (s *CMSService) UpdateAd(id uint, updates *model.Ad) (*model.Ad, error) {
	ad, err := s.adRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if updates.Name != "" {
		ad.Name = updates.Name
	}
	if len(updates.StoreTypes) > 0 {
		ad.StoreTypes = updates.StoreTypes
	}
	ad.PlaylistOrder = updates.PlaylistOrder
	ad.IsActive = updates.IsActive
	if updates.StartDate != nil {
		ad.StartDate = updates.StartDate
	}
	if updates.EndDate != nil {
		ad.EndDate = updates.EndDate
	}
	if updates.StartTime != nil {
		ad.StartTime = updates.StartTime
	}
	if updates.EndTime != nil {
		ad.EndTime = updates.EndTime
	}
	if err := validateSchedule(ad.StartDate, ad.EndDate); err != nil {
		return nil, err
	}
	if err := s.adRepo.Update(ad); err != nil {
		return nil, err
	}
	return ad, nil
}

func (s *CMSService) DeleteAd(id uint) error {
	ad, err := s.adRepo.FindByID(id)
	if err != nil {
		return err
	}
	if err := os.Remove(ad.StoragePath); err != nil && !os.IsNotExist(err) {
		// log but continue
	}
	return s.adRepo.Delete(id)
}

func (s *CMSService) ToggleAd(id uint) (*model.Ad, error) {
	ad, err := s.adRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	ad.IsActive = !ad.IsActive
	if err := s.adRepo.Update(ad); err != nil {
		return nil, err
	}
	return ad, nil
}

func (s *CMSService) GetActivePlaylist(storeType string) ([]model.Ad, error) {
	return s.adRepo.FindActiveByStoreType(storeType)
}

func (s *CMSService) LogPlay(adID uint, terminalID, storeType string) error {
	log := &model.AdPlayLog{
		AdID:       adID,
		TerminalID: terminalID,
		StoreType:  storeType,
	}
	return s.playLogRepo.Create(log)
}

func (s *CMSService) GetAdFilePath(id uint) (string, error) {
	ad, err := s.adRepo.FindByID(id)
	if err != nil {
		return "", err
	}
	return ad.StoragePath, nil
}

func (s *CMSService) ValidateUploadFile(filename string, size int64) error {
	if size > 50*1024*1024 {
		return fmt.Errorf("FILE_TOO_LARGE")
	}
	ext := strings.ToLower(filepath.Ext(filename))
	allowed := map[string]bool{".mp4": true, ".webm": true, ".mov": true}
	if !allowed[ext] {
		return fmt.Errorf("INVALID_FILE_TYPE")
	}
	return nil
}

func ParseTimePointer(t string) *time.Time {
	if t == "" {
		return nil
	}
	for _, layout := range []string{"15:04:05", "15:04"} {
		parsed, err := time.Parse(layout, t)
		if err == nil {
			return &parsed
		}
	}
	return nil
}

func ParseDatePointer(d string) *time.Time {
	if d == "" {
		return nil
	}
	parsed, err := time.Parse("2006-01-02", d)
	if err != nil {
		return nil
	}
	return &parsed
}

func ParseInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}
