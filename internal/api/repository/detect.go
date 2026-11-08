package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Timur1414/asvt/pkg/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type DetectRepository interface {
	SaveNewObject(ctx context.Context, detectedObj DetectedObject) error
	GetAll(ctx context.Context) ([]DetectedObject, error)
	GetObjectByUUid(ctx context.Context, uuid string) (DetectedObject, error)
}

const (
	detectedObjectPrefix = "obj:"
	detectedObjectTTL    = time.Second * 3
)

type DetectRedis struct {
	db *redis.Client
}

func NewDetectRedis(client *redis.Client) *DetectRedis {
	return &DetectRedis{db: client}
}

func (obj *DetectRedis) SaveNewObject(ctx context.Context, detectedObj DetectedObject) error {
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

func (obj *DetectRedis) GetAll(ctx context.Context) ([]DetectedObject, error) {
	log := logger.GetLoggerWithRequestId(ctx)
	res := make([]DetectedObject, 0)
	iter := obj.db.Scan(ctx, 0, detectedObjectPrefix, 0).Iterator()
	if iter.Err() != nil {
		log.Error("fail to iterate detected objects", zap.String("uuid", iter.Val()))
		return []DetectedObject{}, iter.Err()
	}
	for iter.Next(ctx) {
		key := iter.Val()
		data, err := obj.db.Get(ctx, key).Bytes()
		if err != nil {
			log.Error("failed to get detected object", zap.String("key", key), zap.Error(err))
			return []DetectedObject{}, err
		}
		var detectedObj DetectedObject
		err = json.Unmarshal(data, &detectedObj)
		if err != nil {
			log.Error("failed to unmarshal detected object", zap.String("key", key), zap.Error(err))
			return []DetectedObject{}, err
		}
		res = append(res, detectedObj)
	}
	return res, nil
}

func (obj *DetectRedis) GetObjectByUUid(ctx context.Context, uuid string) (DetectedObject, error) {
	log := logger.GetLoggerWithRequestId(ctx)
	key := detectedObjectPrefix + uuid
	data, err := obj.db.Get(ctx, key).Bytes()
	if err != nil {
		log.Error("failed to get detected object", zap.String("uuid", uuid), zap.Error(err))
		return DetectedObject{}, err
	}
	var detectedObj DetectedObject
	err = json.Unmarshal(data, &detectedObj)
	if err != nil {
		log.Error("failed to unmarshal detected object", zap.String("uuid", uuid), zap.Error(err))
		return DetectedObject{}, err
	}
	return detectedObj, nil
}
