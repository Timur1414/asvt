package usecase

import (
	"context"
	"slices"
	"time"

	"github.com/Timur1414/asvt/internal/security_checker/domain"
	"github.com/Timur1414/asvt/internal/security_checker/repository"
)

type EventsUseCase interface {
	GetSuspiciousObjects(ctx context.Context) ([]domain.DetectedObject, error)
}

type Events struct {
	eventsRepository repository.EventsRepository
}

func NewEvents(eventsRepo repository.EventsRepository) *Events {
	return &Events{
		eventsRepository: eventsRepo,
	}
}

func (obj *Events) GetSuspiciousObjects(ctx context.Context) ([]domain.DetectedObject, error) {
	objects, err := obj.eventsRepository.GetNotNoted(ctx)
	if err != nil {
		return []domain.DetectedObject{}, err
	}
	objects = slices.DeleteFunc(objects, func(item domain.DetectedObject) bool {
		duration := time.Since(item.CreatedAt)
		return duration < time.Second*15
	})
	return objects, nil
}
