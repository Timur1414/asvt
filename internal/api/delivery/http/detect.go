package http

import (
	"encoding/json"
	"net/http"

	"github.com/Timur1414/asvt/pkg/logger"
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
	w.WriteHeader(http.StatusNoContent)
}
