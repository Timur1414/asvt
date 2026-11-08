package usecase

import (
	"context"

	"github.com/Timur1414/asvt/internal/api/repository"
	"github.com/google/uuid"
)

type DetectUseCase interface {
	MatchOrCreate(ctx context.Context, detectedObj repository.DetectedObject) error
	// ToDo get nearest object (by Bbox)
	// ToDo save new objects + create uuid
}

type Detect struct {
	detectRepository repository.DetectRepository
}

func NewDetect(detectRepo repository.DetectRepository) *Detect {
	return &Detect{
		detectRepository: detectRepo,
	}
}

func (obj *Detect) MatchOrCreate(ctx context.Context, detectedObj repository.DetectedObject) error {
	existingObjects, err := obj.detectRepository.GetAll(ctx)
	if err != nil {
		return err
	}
	uuid, found := MatchOrCreate(detectedObj, existingObjects, 0.5)
	if found {
		return nil
	}
}

func MatchOrCreate(incoming repository.DetectedObject, existingObjects []repository.DetectedObject, iouThreshold float64) (string, bool) {
	var bestUUID string
	maxIoU := 0.0
	for _, obj := range existingObjects {
		iou := repository.CalculateIoU(incoming, obj)
		if iou > maxIoU {
			maxIoU = iou
			bestUUID = obj.Uuid
		}
	}
	// Если IoU выше порога (например, 0.4 - 0.5), считаем, что это тот же объект
	if maxIoU >= iouThreshold {
		return bestUUID, true
	}
	return "", false
}
