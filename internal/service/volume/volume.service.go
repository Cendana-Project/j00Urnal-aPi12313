package volume

import (
	"context"

	"github.com/api-monolith-template/internal/constant"
	"github.com/api-monolith-template/internal/model/entity"
	"github.com/api-monolith-template/internal/repository/journal"
	"github.com/api-monolith-template/internal/repository/volume"
	"github.com/api-monolith-template/internal/service/issue"
)

type Service struct {
	volumeRepo   *volume.Repository
	journalRepo  *journal.Repository
	issueService *issue.Service
}

func NewService(vr *volume.Repository, jr *journal.Repository, is *issue.Service) *Service {
	return &Service{
		volumeRepo:   vr,
		journalRepo:  jr,
		issueService: is,
	}
}

func (s *Service) Create(ctx context.Context, journalID string, year, number int) (*entity.Volume, error) {
	// Check Journal existence
	j, err := s.journalRepo.GetByID(ctx, journalID)
	if err != nil {
		return nil, err
	}
	if j == nil {
		return nil, constant.ErrRecordNotFound
	}

	// Check for duplicates
	exists, err := s.volumeRepo.Exists(ctx, journalID, year, number)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, constant.ErrConflict // Use generic conflict or create specific
	}

	vol := &entity.Volume{
		JournalID: journalID,
		Year:      year,
		Number:    number,
		Status:    constant.PublicationStatusDraft,
	}

	if err := s.volumeRepo.Create(ctx, vol); err != nil {
		return nil, err
	}
	return vol, nil
}

func (s *Service) Update(ctx context.Context, id string, year, number int) (*entity.Volume, error) {
	vol, err := s.volumeRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if vol == nil {
		return nil, constant.ErrRecordNotFound
	}

	// Check for duplicates if changed
	if vol.Year != year || vol.Number != number {
		exists, err := s.volumeRepo.Exists(ctx, vol.JournalID, year, number)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, constant.ErrConflict
		}
	}

	vol.Year = year
	vol.Number = number
	if err := s.volumeRepo.Update(ctx, vol); err != nil {
		return nil, err
	}
	return vol, nil
}

func (s *Service) SetStatus(ctx context.Context, id string, status constant.PublicationStatus) (*entity.Volume, error) {
	vol, err := s.volumeRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if vol == nil {
		return nil, constant.ErrRecordNotFound
	}

	vol.Status = status
	if err := s.volumeRepo.Update(ctx, vol); err != nil {
		return nil, err
	}
	return vol, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	// 1. Get Volume
	vol, err := s.volumeRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if vol == nil {
		return constant.ErrRecordNotFound
	}

	// 2. Cascade Delete Issues
	for _, iss := range vol.Issues {
		if err := s.issueService.Delete(ctx, iss.ID); err != nil {
			return err
		}
	}

	// 3. Delete Volume Record
	return s.volumeRepo.Delete(ctx, id)
}

func (s *Service) GetAll(ctx context.Context) ([]*entity.Volume, error) {
	return s.volumeRepo.FindAll(ctx)
}

func (s *Service) GetByID(ctx context.Context, id string) (*entity.Volume, error) {
	return s.volumeRepo.GetByID(ctx, id)
}
