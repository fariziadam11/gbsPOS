package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"gbs-pos-api/internal/config"
	"gbs-pos-api/internal/database"
	"gbs-pos-api/internal/handler"
	"gbs-pos-api/internal/model"
	"gbs-pos-api/internal/repository"
	"gbs-pos-api/internal/router"
	"gbs-pos-api/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("failed to load config: ", err)
	}

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	level, _ := zerolog.ParseLevel(cfg.LogLevel)
	zerolog.SetGlobalLevel(level)

	db, err := database.Connect(cfg.DatabaseURL, cfg.LogLevel)
	if err != nil {
		log.Fatal("failed to connect database: ", err)
	}

	if cfg.MigrationsPath != "" {
		if err := database.RunMigrations(cfg.DatabaseURL, cfg.MigrationsPath); err != nil {
			log.Fatal("failed to run migrations: ", err)
		}
	} else {
		if err := database.Migrate(db,
			&model.User{},
			&model.Product{},
			&model.Order{},
			&model.OrderItem{},
			&model.Settlement{},
		); err != nil {
			log.Fatal("failed to migrate database: ", err)
		}
	}

	database.Seed(db)

	userRepo := repository.NewUserRepository(db)
	productRepo := repository.NewProductRepository(db)
	orderRepo := repository.NewOrderRepository(db)
	settlementRepo := repository.NewSettlementRepository(db)

	authService := service.NewAuthService(userRepo, cfg.JWTSecret, cfg.JWTExpiryHours)
	productService := service.NewProductService(productRepo)
	orderService := service.NewOrderService(orderRepo)
	settlementService := service.NewSettlementService(orderRepo, settlementRepo)

	authHandler := handler.NewAuthHandler(authService)
	productHandler := handler.NewProductHandler(productService)
	orderHandler := handler.NewOrderHandler(orderService, settlementService)
	settlementHandler := handler.NewSettlementHandler(settlementService)

	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := router.Setup(
		cfg,
		authHandler,
		productHandler,
		orderHandler,
		settlementHandler,
	)

	srv := &http.Server{
		Addr:           ":" + cfg.Port,
		Handler:        r,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		IdleTimeout:    120 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Println("POS API starting on port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("server failed: ", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down POS API...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatal("forced shutdown: ", err)
	}
	log.Println("POS API stopped")
}
