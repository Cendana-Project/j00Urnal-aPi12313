package term

import (
	"context"

	"github.com/api-monolith-template/internal/model/entity"
	"github.com/api-monolith-template/internal/repository/term"
)

type Service struct {
	termRepo *term.Repository
}

func NewService(tr *term.Repository) *Service {
	return &Service{termRepo: tr}
}

func (s *Service) GetActive(ctx context.Context) (*entity.PublicationTerm, error) {
	return s.termRepo.GetActive(ctx)
}

func (s *Service) Create(ctx context.Context, content string) (*entity.PublicationTerm, error) {
	return s.termRepo.Create(ctx, content)
}
