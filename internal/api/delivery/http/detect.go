package http

import (
	"encoding/json"
	"net/http"

	"github.com/Timur1414/asvt/pkg/logger"
	"github.com/Timur1414/asvt/pkg/web_helpers"
	"go.uber.org/zap"
)

type Detection struct {
}

func NewDetection() *Detection {
	return &Detection{}
}

func (obj *Detection) Detect(w http.ResponseWriter, r *http.Request) {
	log := logger.GetLoggerWithRequestId(r.Context())
	var request DetectRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	defer func() {
		err = r.Body.Close()
		if err != nil {
			log.Error("failed to close request body", zap.Error(err))
		}
	}()
	if err != nil {
		log.Error("Error decoding request", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Info("DetectRequest received", zap.Any("request", request))
	// ToDo get objects - check in memory - create ids
	// ToDo new people - create event
	// ToDo small TTL in redis
	// ToDo people is ge 15 s - SMTP
	response := web_helpers.NewOkResponse()
	web_helpers.WriteResponseJSON(w, response.Code, response)
}
