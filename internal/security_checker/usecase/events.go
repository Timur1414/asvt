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
	GetPostRollObjects(ctx context.Context, detectedObjects []domain.DetectedObject) ([]domain.DetectedObject, time.Time, time.Time, error)
	SaveNewEvent(ctx context.Context, object domain.DetectedObject) error
	SaveObjects(ctx context.Context, objects []domain.DetectedObject) error
	DeleteObjects(ctx context.Context, objects []domain.DetectedObject) error
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
	objects, err := obj.objectsRepository.GetAll(ctx)
	if err != nil {
		return []domain.DetectedObject{}, err
	}
	objects = slices.DeleteFunc(objects, func(item domain.DetectedObject) bool {
		return !(item.Status == domain.StatusSuspicious || item.Status == domain.StatusWaitingPostRoll)
	})
	return objects, nil
}

func (obj *Events) NoteObject(ctx context.Context, detectedObject domain.DetectedObject) error {
	if detectedObject.Status != domain.StatusSuspicious {
		return nil
	}
	if time.Since(detectedObject.LastSeenAt) <= 2*time.Second {
		return nil
	}
	err := obj.objectsRepository.Delete(ctx, detectedObject.Uuid)
	if err != nil {
		return err
	}
	detectedObject.Status = domain.StatusWaitingPostRoll
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

func (obj *Events) GetPostRollObjects(ctx context.Context, detectedObjects []domain.DetectedObject) ([]domain.DetectedObject, time.Time, time.Time, error) {
	res := make([]domain.DetectedObject, 0)
	var startFragmentTime time.Time
	var endFragmentTime time.Time
	for _, detectedObject := range detectedObjects {
		if detectedObject.Status == domain.StatusWaitingPostRoll && time.Since(detectedObject.LastSeenAt) >= 5*time.Second {
			detectedObject.Status = domain.StatusCompleted
			res = append(res, detectedObject)
			if startFragmentTime.IsZero() || detectedObject.CreatedAt.Before(startFragmentTime) {
				startFragmentTime = detectedObject.CreatedAt
			}
			if endFragmentTime.IsZero() || detectedObject.LastSeenAt.After(endFragmentTime) {
				endFragmentTime = detectedObject.LastSeenAt
			}
			err := obj.objectsRepository.Delete(ctx, detectedObject.Uuid)
			if err != nil {
				return res, startFragmentTime, endFragmentTime, err
			}
			err = obj.objectsRepository.SaveNewObject(ctx, detectedObject)
			if err != nil {
				return res, startFragmentTime, endFragmentTime, err
			}
		}
	}
	if len(res) == 0 {
		return res, startFragmentTime, endFragmentTime, nil
	}
	startFragmentTime = startFragmentTime.Add(-5 * time.Second)
	return res, startFragmentTime, endFragmentTime, nil
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

func (obj *Events) DeleteObjects(ctx context.Context, objects []domain.DetectedObject) error {
	for _, object := range objects {
		err := obj.objectsRepository.Delete(ctx, object.Uuid)
		if err != nil {
			return err
		}
	}
	return nil
}
