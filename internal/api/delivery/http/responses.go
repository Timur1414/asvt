package http

import (
	"time"

	"github.com/Timur1414/asvt/internal/api/domain"
)

type EventResponse struct {
	ObjectType       string    `json:"object_type"`
	ObjectConfidence float32   `json:"object_confidence"`
	CreatedAt        time.Time `json:"created_at"`
}

type HistoryEventsResponse struct {
	Events []EventResponse `json:"events"`
}

func NewHistoryEventsResponse(events []domain.SecurityEvent) HistoryEventsResponse {
	eventsResponses := make([]EventResponse, len(events))
	for i, event := range events {
		eventsResponses[i] = EventResponse{
			ObjectType:       event.ObjectType,
			ObjectConfidence: event.ObjectConfidence,
			CreatedAt:        event.CreatedAt,
		}
	}
	return HistoryEventsResponse{
		Events: eventsResponses,
	}
}
