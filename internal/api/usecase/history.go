package usecase

import (
	"context"

	"github.com/Timur1414/asvt/internal/api/domain"
	"github.com/Timur1414/asvt/internal/api/repository"
)

type HistoryUseCase interface {
	GetAll(ctx context.Context) ([]domain.SecurityEvent, error)
}

type History struct {
	repo repository.HistoryRepository
}

func NewHistory(repo repository.HistoryRepository) *History {
	return &History{
		repo: repo,
	}
}

func (obj *History) GetAll(ctx context.Context) ([]domain.SecurityEvent, error) {
	return obj.repo.GetAll(ctx)
}
