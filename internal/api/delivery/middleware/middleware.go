package middleware

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/Timur1414/asvt/pkg/context_helper"
	"github.com/Timur1414/asvt/pkg/logger"
	"github.com/Timur1414/asvt/pkg/web_helpers"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type responseWriter struct {
	http.ResponseWriter
	StatusCode int
}

func (obj *responseWriter) WriteHeader(code int) {
	obj.StatusCode = code
	obj.ResponseWriter.WriteHeader(code)
}

func (obj *responseWriter) Write(b []byte) (int, error) {
	if obj.StatusCode == 0 {
		obj.StatusCode = http.StatusOK
	}
	return obj.ResponseWriter.Write(b)
}

func PanicMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := logger.GetLogger()
		defer func() {
			if err := recover(); err != nil {
				requestId, ok := r.Context().Value(context_helper.ContextKeyRequestId).(string)
				if !ok {
					log.Error("[panic middleware] panic",
						zap.Any("err", err))
				} else {
					log.Error("[panic middleware] panic",
						zap.String("request_id", requestId),
						zap.Any("err", err))
				}
				response := web_helpers.NewServerErrorResponse(requestId)
				web_helpers.WriteResponseJSON(w, response.Code, response)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func AccessLogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := logger.GetAccessLogger()
		requestId := uuid.New().String()
		ctx := context.WithValue(r.Context(), context_helper.ContextKeyRequestId, requestId)
		log.Info("Request",
			zap.String("path", r.URL.Path),
			zap.String("method", r.Method),
			zap.String("remote_addr", r.RemoteAddr),
			zap.String("user_agent", r.UserAgent()),
			zap.String("request_id", requestId))
		wr := &responseWriter{
			ResponseWriter: w,
			StatusCode:     http.StatusOK,
		}
		timeStart := time.Now()
		next.ServeHTTP(wr, r.WithContext(ctx))
		duration := time.Since(timeStart)
		log.Info("Response",
			zap.String("path", r.URL.Path),
			zap.String("method", r.Method),
			zap.String("remote_addr", r.RemoteAddr),
			zap.String("request_id", requestId),
			zap.Int("status_code", wr.StatusCode),
			zap.String("duration", duration.String()))
	})
}

func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", os.Getenv("FRONT_URL"))
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Origin, Cache-Control, X-Requested-With")
		if r.Method == http.MethodOptions {
			return
		}
		next.ServeHTTP(w, r)
	})
}
