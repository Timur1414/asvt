package http

import (
	"net/http"

	"github.com/Timur1414/asvt/internal/api/usecase"
	"github.com/Timur1414/asvt/pkg/context_helper"
	"github.com/Timur1414/asvt/pkg/logger"
	"github.com/Timur1414/asvt/pkg/web_helpers"
)

type HistoryHandler struct {
	uc usecase.HistoryUseCase
}

func NewHistoryHandler(uc usecase.HistoryUseCase) *HistoryHandler {
	return &HistoryHandler{uc: uc}
}

func (obj *HistoryHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	log := logger.GetLoggerWithRequestId(r.Context())
	log.Info("Get events request")
	requestId := context_helper.GetRequestIdFromContext(r.Context())
	events, err := obj.uc.GetAll(r.Context())
	if err != nil {
		response := web_helpers.NewServerErrorResponse(requestId)
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}
	response := NewHistoryEventsResponse(events)
	web_helpers.WriteResponseJSON(w, http.StatusOK, response)
}
