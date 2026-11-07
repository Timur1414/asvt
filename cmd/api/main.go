package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/Timur1414/asvt/internal/api/delivery/middleware"
	"github.com/Timur1414/asvt/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	DEBUG := os.Getenv("DEBUG") == "true"
	err := logger.InitLogger(DEBUG)
	if err != nil {
		fmt.Println("Error initializing logger: ", err)
		return
	}
	defer func() {
		err = logger.Close()
		if err != nil {
			fmt.Println("Error closing logger: ", err)
		}
	}()
	log := logger.GetLogger()
	err = logger.InitAccessLogger()
	if err != nil {
		log.Error("Error initializing access logger", zap.Error(err))
		return
	}
	defer func() {
		err = logger.AccessClose()
		if err != nil {
			fmt.Println("Error closing access logger: ", err)
		}
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("POST /api/detection", func(w http.ResponseWriter, r *http.Request) {
		type BBox struct {
			X      int `json:"x"`
			Y      int `json:"y"`
			Width  int `json:"width"`
			Height int `json:"height"`
		}
		type Object struct {
			ClassId    int     `json:"class_id"`
			ClassName  string  `json:"class_name"`
			Confidence float32 `json:"confidence"`
			Bbox       BBox    `json:"bbox"`
		}
		type Request struct {
			DeviceID    string   `json:"device_id"`
			Timestamp   int64    `json:"timestamp"`
			TimestampMs string   `json:"timestamp_ms"`
			FrameId     int      `json:"frame_id"`
			FrameWidth  int      `json:"frame_width"`
			FrameHeight int      `json:"frame_height"`
			Objects     []Object `json:"objects"`
		}
		var request Request
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
		log.Info("Request received", zap.Any("request", request))
		w.WriteHeader(http.StatusNoContent)
	})

	handler := middleware.CORSMiddleware(mux)
	handler = middleware.PanicMiddleware(handler)
	handler = middleware.AccessLogMiddleware(handler)

	addr := ":8080"
	server := http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 120 * time.Second,
	}
	log.Info("starting api server",
		zap.String("addr", addr),
	)
	err = server.ListenAndServe()
	if err != nil {
		log.Error("Error starting server", zap.Error(err))
		return
	}
}
