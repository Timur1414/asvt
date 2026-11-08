package repository

import (
	"context"
	"encoding/json"
	"slices"
	"time"

	"github.com/Timur1414/asvt/internal/security_checker/domain"
	"github.com/Timur1414/asvt/pkg/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type EventsRepository interface {
	GetAll(ctx context.Context) ([]domain.DetectedObject, error)
	GetNotNoted(ctx context.Context) ([]domain.DetectedObject, error)
	SaveNewObject(ctx context.Context, detectedObj domain.DetectedObject) error
	Delete(ctx context.Context, uuid string) error
}

const (
	detectedObjectPrefix = "obj:"
	detectedObjectTTL    = time.Second * 3
)

type EventsRedis struct {
	db *redis.Client
}

func NewEventsRedis(client *redis.Client) *EventsRedis {
	return &EventsRedis{
		db: client,
	}
}

func (obj *EventsRedis) GetAll(ctx context.Context) ([]domain.DetectedObject, error) {
	log := logger.GetLogger()
	res := make([]domain.DetectedObject, 0)
	iter := obj.db.Scan(ctx, 0, detectedObjectPrefix, 0).Iterator()
	if iter.Err() != nil {
		log.Error("fail to iterate detected objects", zap.String("uuid", iter.Val()))
		return []domain.DetectedObject{}, iter.Err()
	}
	for iter.Next(ctx) {
		key := iter.Val()
		data, err := obj.db.Get(ctx, key).Bytes()
		if err != nil {
			log.Error("failed to get detected object", zap.String("key", key), zap.Error(err))
			return []domain.DetectedObject{}, err
		}
		var detectedObj domain.DetectedObject
		err = json.Unmarshal(data, &detectedObj)
		if err != nil {
			log.Error("failed to unmarshal detected object", zap.String("key", key), zap.Error(err))
			return []domain.DetectedObject{}, err
		}
		res = append(res, detectedObj)
	}
	return res, nil
}

func (obj *EventsRedis) GetNotNoted(ctx context.Context) ([]domain.DetectedObject, error) {
	log := logger.GetLogger()
	objects, err := obj.GetAll(ctx)
	if err != nil {
		return []domain.DetectedObject{}, err
	}
	objects = slices.DeleteFunc(objects, func(item domain.DetectedObject) bool {
		return item.Noted
	})
	log.Info("deleted noted objects")
	return objects, nil
}

func (obj *EventsRedis) SaveNewObject(ctx context.Context, detectedObj domain.DetectedObject) error {
	log := logger.GetLoggerWithRequestId(ctx)
	key := detectedObjectPrefix + detectedObj.Uuid
	data, err := json.Marshal(detectedObj)
	if err != nil {
		log.Error("fail to marshal detected object", zap.Error(err))
		return err
	}
	err = obj.db.Set(ctx, key, data, detectedObjectTTL).Err()
	if err != nil {
		log.Error("failed to save detected object", zap.String("uuid", detectedObj.Uuid), zap.Error(err))
		return err
	}
	log.Info("saved detected object", zap.String("uuid", detectedObj.Uuid))
	return nil
}

func (obj *EventsRedis) Delete(ctx context.Context, uuid string) error {
	log := logger.GetLoggerWithRequestId(ctx)
	key := detectedObjectPrefix + uuid
	err := obj.db.Del(ctx, key).Err()
	if err != nil {
		log.Error("failed to delete detected object", zap.String("uuid", uuid), zap.Error(err))
		return err
	}
	log.Info("deleted detected object", zap.String("uuid", uuid))
	return nil
}
