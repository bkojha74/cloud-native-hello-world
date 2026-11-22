package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"backend/internal/cache"
	"backend/internal/config"
	"backend/internal/handler"
	"backend/internal/middleware"
	"backend/internal/repository"
	"backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	cfg := config.Load()

	if cfg.LogLevel != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 1. Initialize MongoDB Client
	mongoCtx, mongoCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer mongoCancel()

	mongoClient, err := mongo.Connect(mongoCtx, options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		log.Printf("[WARN] Failed to create MongoDB client: %v", err)
	} else {
		if err := mongoClient.Ping(mongoCtx, nil); err != nil {
			log.Printf("[WARN] MongoDB ping failed: %v", err)
		} else {
			log.Println("[INFO] Successfully connected to MongoDB")
		}
	}

	mongoDB := mongoClient.Database(cfg.MongoDB)
	mongoCollection := mongoDB.Collection("messages")
	mongoRepo := repository.NewMongoMessageRepository(mongoCollection)

	// Ensure seed data
	seedCtx, seedCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer seedCancel()
	if err := mongoRepo.EnsureSeedData(seedCtx); err != nil {
		log.Printf("[WARN] MongoDB ensure seed data error: %v", err)
	}

	// 2. Initialize Redis Client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
	})
	redisCtx, redisCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer redisCancel()
	if err := redisClient.Ping(redisCtx).Err(); err != nil {
		log.Printf("[WARN] Redis ping failed: %v", err)
	} else {
		log.Println("[INFO] Successfully connected to Redis")
	}

	messageCache := cache.NewRedisMessageCache(redisClient, time.Duration(cfg.RedisTTLSeconds)*time.Second)

	// 3. Service & Handlers
	msgService := service.NewMessageService(mongoRepo, messageCache)
	msgHandler := handler.NewMessageHandler(msgService)
	healthHandler := handler.NewHealthHandler()

	// 4. Setup Router & Middleware
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.RequestID())
	router.Use(middleware.StructuredLogger())
	router.Use(middleware.PrometheusMetrics())

	// Simple CORS middleware for local frontend dev
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Request-ID")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	// Routes
	router.GET("/message", msgHandler.GetMessage)
	router.GET("/health", healthHandler.GetHealth)
	router.HEAD("/health", healthHandler.GetHealth)
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// 5. Start Server with Graceful Shutdown
	srv := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: router,
	}

	go func() {
		log.Printf("[INFO] Server listening on %s", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[FATAL] Server listen error: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	log.Printf("[INFO] Received signal: %v. Initiating graceful shutdown...", sig)

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Duration(cfg.ShutdownTimeoutSeconds)*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("[ERROR] Server forced to shutdown: %v", err)
	}

	if err := mongoClient.Disconnect(shutdownCtx); err != nil {
		log.Printf("[WARN] MongoDB disconnect error: %v", err)
	}

	if err := redisClient.Close(); err != nil {
		log.Printf("[WARN] Redis close error: %v", err)
	}

	log.Println("[INFO] Server exited successfully.")
}
