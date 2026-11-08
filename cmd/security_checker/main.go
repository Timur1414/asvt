package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/Timur1414/asvt/internal/security_checker/repository"
	"github.com/Timur1414/asvt/internal/security_checker/usecase"
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

	eventsRepo := repository.NewEventsRedis(redisClient)
	log.Info("Repository initialized")
	eventsUsecase := usecase.NewEvents(eventsRepo)
	log.Info("Usecase initialized")

	interval := time.Second * 30
	checkCtx := context.Background()
	for {
		suspiciousObjects, err := eventsUsecase.GetSuspiciousObjects(checkCtx)
		if err != nil {
			log.Error("Failed to get suspicious objects", zap.Error(err))
			time.Sleep(interval)
			continue
		}
		err = eventsUsecase.NoteObjects(checkCtx, suspiciousObjects)
		if err != nil {
			log.Error("Failed to note suspicious objects", zap.Error(err))
			time.Sleep(interval)
			continue
		}
		// ToDo deliver suspiciousObjects
		select {
		case <-checkCtx.Done():
			log.Info("Context cancelled, terminating aggregator service")
			return
		case <-time.After(interval):
		}
	}
}
