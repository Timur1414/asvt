package repository

import (
	"context"
	"time"

	"github.com/Timur1414/asvt/internal/api/domain"
	"github.com/Timur1414/asvt/pkg/logger"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type HistoryRepository interface {
	GetAll(ctx context.Context) ([]domain.SecurityEvent, error)
}

type HistoryPostgres struct {
	db *pgxpool.Pool
}

func NewHistoryPostgres(db *pgxpool.Pool) *HistoryPostgres {
	return &HistoryPostgres{db: db}
}

func (obj *HistoryPostgres) GetAll(ctx context.Context) ([]domain.SecurityEvent, error) {
	log := logger.GetLoggerWithRequestId(ctx)
	query := `select id, object_uuid, object_type, object_confidence, created_at from security_event order by created_at desc;`
	var args []any
	start := time.Now()
	rows, err := obj.db.Query(ctx, query, args...)
	if err != nil {
		log.Error("failed to get security events", zap.Error(err))
		return []domain.SecurityEvent{}, err
	}
	defer rows.Close()
	events := make([]domain.SecurityEvent, 0)
	for rows.Next() {
		var event domain.SecurityEvent
		err = rows.Scan(&event.Id, &event.ObjectUuid, &event.ObjectType, &event.ObjectConfidence, &event.CreatedAt)
		if err != nil {
			log.Error("failed to scan row", zap.Error(err))
			return []domain.SecurityEvent{}, err
		}
		events = append(events, event)
	}
	duration := time.Since(start)
	log = logger.ModifyLoggerWithDBQuery(log, query, args, duration)
	log.Info("Query executed")
	return events, nil
}
