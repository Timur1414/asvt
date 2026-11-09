package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	delivery "github.com/Timur1414/asvt/internal/api/delivery/http"
	"github.com/Timur1414/asvt/internal/api/delivery/middleware"
	"github.com/Timur1414/asvt/internal/api/repository"
	"github.com/Timur1414/asvt/internal/api/usecase"
	"github.com/Timur1414/asvt/pkg/cache"
	"github.com/Timur1414/asvt/pkg/database"
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

	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_DB"),
	)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	dbPool, err := database.New(ctx, connStr)
	if err != nil {
		log.Fatal("Failed to connect to postgres", zap.Error(err))
	}
	defer dbPool.Close()
	log.Info("Connected to Postgres")

	redisAddr := fmt.Sprintf("redis://:%s@%s:%s/0", os.Getenv("REDIS_PASSWORD"), os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT"))
	redisClient, err := cache.New(ctx, redisAddr)
	if err != nil {
		log.Fatal("Failed to connect to redis", zap.Error(err))
	}
	defer func() {
		err = redisClient.Close()
		if err != nil {
			log.Fatal("Failed to close redis connection", zap.Error(err))
		}
	}()
	log.Info("Connected to Redis")

	detectRepo := repository.NewDetectRedis(redisClient)
	historyRepo := repository.NewHistoryPostgres(dbPool)
	log.Info("Repository initialized")
	detectUsecase := usecase.NewDetect(detectRepo)
	historyUsecase := usecase.NewHistory(historyRepo)
	log.Info("Usecase initialized")
	detectionHandler := delivery.NewDetection(detectUsecase)
	historyHandler := delivery.NewHistoryHandler(historyUsecase)
	log.Info("Handler initialized")

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("POST /api/detection", detectionHandler.Detect)
	mux.HandleFunc("GET /api/history", historyHandler.GetAll)

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
