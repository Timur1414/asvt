package http

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/Timur1414/asvt/internal/api/domain"
	"github.com/Timur1414/asvt/internal/api/usecase"
	"github.com/Timur1414/asvt/pkg/logger"
	"github.com/Timur1414/asvt/pkg/web_helpers"
	"go.uber.org/zap"
)

type Detection struct {
	detectUsecase usecase.DetectUseCase
	mode          bool
	mu            sync.RWMutex
}

func NewDetection(detectUC usecase.DetectUseCase) *Detection {
	return &Detection{
		detectUsecase: detectUC,
		mode:          false,
	}
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
	mode := obj.getMode()
	if mode {
		hasCar := false
		for _, detectedObject := range request.Objects {
			if detectedObject.ClassName == "car" {
				hasCar = true
			}
		}
		if !hasCar {
			response := web_helpers.NewOkResponse()
			web_helpers.WriteResponseJSON(w, response.Code, response)
			return
		}
		for _, detectedObj := range request.Objects {
			if detectedObj.ClassName == "person" {
				_, err = obj.detectUsecase.MatchOrCreate(r.Context(), domain.DetectedObject{
					Uuid:       "",
					ClassId:    detectedObj.ClassId,
					ClassName:  detectedObj.ClassName,
					Confidence: detectedObj.Confidence,
					X:          detectedObj.Bbox.X,
					Y:          detectedObj.Bbox.Y,
					Width:      detectedObj.Bbox.Width,
					Height:     detectedObj.Bbox.Height,
					CreatedAt:  time.Time{},
				})
			}
		}
	} else {
		for _, detectedObj := range request.Objects {
			if detectedObj.ClassName == "person" || detectedObj.ClassName == "car" {
				_, err = obj.detectUsecase.MatchOrCreate(r.Context(), domain.DetectedObject{
					Uuid:       "",
					ClassId:    detectedObj.ClassId,
					ClassName:  detectedObj.ClassName,
					Confidence: detectedObj.Confidence,
					X:          detectedObj.Bbox.X,
					Y:          detectedObj.Bbox.Y,
					Width:      detectedObj.Bbox.Width,
					Height:     detectedObj.Bbox.Height,
					CreatedAt:  time.Time{},
				})
			}
		}
	}
	response := web_helpers.NewOkResponse()
	web_helpers.WriteResponseJSON(w, response.Code, response)
}

func (obj *Detection) ToggleMode(w http.ResponseWriter, r *http.Request) {
	log := logger.GetLoggerWithRequestId(r.Context())
	var request ToggleModeRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	defer r.Body.Close()
	if err != nil {
		log.Error("Error decoding request", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	obj.setMode(request.Mode)
	log.Info("Get toggle mode request", zap.Any("External mode", request.Mode))
}

func (obj *Detection) getMode() bool {
	obj.mu.RLock()
	defer obj.mu.RUnlock()
	return obj.mode
}

func (obj *Detection) setMode(newMode bool) {
	obj.mu.Lock()
	defer obj.mu.Unlock()
	obj.mode = newMode
}
