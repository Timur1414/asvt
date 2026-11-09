package usecase

import (
	"context"
	"time"

	"github.com/Timur1414/asvt/internal/api/domain"
	"github.com/Timur1414/asvt/internal/api/repository"
	"github.com/Timur1414/asvt/pkg/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type DetectUseCase interface {
	MatchOrCreate(ctx context.Context, detectedObj domain.DetectedObject) (domain.DetectedObject, error)
}

type Detect struct {
	detectRepository repository.DetectRepository
}

func NewDetect(detectRepo repository.DetectRepository) *Detect {
	return &Detect{
		detectRepository: detectRepo,
	}
}

func (obj *Detect) MatchOrCreate(ctx context.Context, detectedObj domain.DetectedObject) (domain.DetectedObject, error) {
	log := logger.GetLoggerWithRequestId(ctx)
	existingObjects, err := obj.detectRepository.GetAll(ctx)
	log.Info("Get existing objects", zap.Any("objects", existingObjects))
	if err != nil {
		return domain.DetectedObject{}, err
	}
	res, found := FindObjectByIoU(detectedObj, existingObjects, 0.5)
	if found {
		log.Info("found object", zap.String("uuid", res.Uuid))
		err = obj.detectRepository.Delete(ctx, res.Uuid)
		if err != nil {
			return domain.DetectedObject{}, err
		}
		err = obj.detectRepository.SaveNewObject(ctx, res)
		if err != nil {
			return domain.DetectedObject{}, err
		}
		return res, nil
	}
	log.Info("object not fount -> creating")
	objUuid := uuid.New().String()
	newObject := domain.DetectedObject{
		Uuid:       objUuid,
		ClassId:    detectedObj.ClassId,
		ClassName:  detectedObj.ClassName,
		Confidence: detectedObj.Confidence,
		X:          detectedObj.X,
		Y:          detectedObj.Y,
		Width:      detectedObj.Width,
		Height:     detectedObj.Height,
		CreatedAt:  time.Now().UTC(),
		Noted:      false,
	}
	err = obj.detectRepository.SaveNewObject(ctx, newObject)
	if err != nil {
		return domain.DetectedObject{}, err
	}
	log.Info("object created", zap.String("uuid", objUuid))
	return newObject, nil
}

func FindObjectByIoU(incoming domain.DetectedObject, existingObjects []domain.DetectedObject, iouThreshold float64) (domain.DetectedObject, bool) {
	log := logger.GetLogger()
	var res domain.DetectedObject
	maxIoU := 0.0
	for _, obj := range existingObjects {
		iou := domain.CalculateIoU(incoming, obj)
		log.Info("Calculated IoU", zap.Float64("IoU", iou))
		if iou > maxIoU {
			maxIoU = iou
			res = obj
		}
	}
	if maxIoU >= iouThreshold {
		return res, true
	}
	return domain.DetectedObject{}, false
}
