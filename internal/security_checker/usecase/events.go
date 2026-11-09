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
	NoteObject(ctx context.Context, detectedObject domain.DetectedObject) error
	NoteObjects(ctx context.Context, detectedObjects []domain.DetectedObject) error
	SaveNewEvent(ctx context.Context, object domain.DetectedObject) error
	SaveObjects(ctx context.Context, objects []domain.DetectedObject) error
}

type Events struct {
	objectsRepository repository.ObjectsRepository
	eventsRepository  repository.EventsRepository
}

func NewEvents(objectsRepo repository.ObjectsRepository, eventsRepo repository.EventsRepository) *Events {
	return &Events{
		objectsRepository: objectsRepo,
		eventsRepository:  eventsRepo,
	}
}

func (obj *Events) GetSuspiciousObjects(ctx context.Context) ([]domain.DetectedObject, error) {
	objects, err := obj.objectsRepository.GetNotNoted(ctx)
	if err != nil {
		return []domain.DetectedObject{}, err
	}
	objects = slices.DeleteFunc(objects, func(item domain.DetectedObject) bool {
		duration := time.Since(item.CreatedAt)
		return duration < time.Second*15
	})
	return objects, nil
}

func (obj *Events) NoteObject(ctx context.Context, detectedObject domain.DetectedObject) error {
	err := obj.objectsRepository.Delete(ctx, detectedObject.Uuid)
	if err != nil {
		return err
	}
	detectedObject.Noted = true
	err = obj.objectsRepository.SaveNewObject(ctx, detectedObject)
	return err
}

func (obj *Events) NoteObjects(ctx context.Context, detectedObjects []domain.DetectedObject) error {
	for _, detectedObject := range detectedObjects {
		err := obj.NoteObject(ctx, detectedObject)
		if err != nil {
			return err
		}
	}
	return nil
}

func (obj *Events) SaveNewEvent(ctx context.Context, object domain.DetectedObject) error {
	event := domain.SecurityEvent{
		ObjectUuid:       object.Uuid,
		ObjectType:       object.ClassName,
		ObjectConfidence: object.Confidence,
		CreatedAt:        time.Now().UTC(),
	}
	return obj.eventsRepository.Save(ctx, event)
}

func (obj *Events) SaveObjects(ctx context.Context, objects []domain.DetectedObject) error {
	for _, object := range objects {
		err := obj.SaveNewEvent(ctx, object)
		if err != nil {
			return err
		}
	}
	return nil
}
