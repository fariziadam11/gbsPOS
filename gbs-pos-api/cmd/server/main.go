package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"gbs-pos-api/internal/config"
	"gbs-pos-api/internal/database"
	"gbs-pos-api/internal/handler"
	"gbs-pos-api/internal/middleware"
	"gbs-pos-api/internal/model"
	"gbs-pos-api/internal/repository"
	"gbs-pos-api/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("failed to load config: ", err)
	}

	os.Setenv("JWT_SECRET", cfg.JWTSecret)
	os.Setenv("JWT_EXPIRY_HOURS", strconv.Itoa(cfg.JWTExpiryHours))

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	level, _ := zerolog.ParseLevel(cfg.LogLevel)
	zerolog.SetGlobalLevel(level)

	db, err := database.Connect(cfg.DatabaseURL, cfg.LogLevel)
	if err != nil {
		log.Fatal("failed to connect database: ", err)
	}

	if err := database.Migrate(db,
		&model.User{},
		&model.Product{},
		&model.Order{},
		&model.OrderItem{},
		&model.Settlement{},
	); err != nil {
		log.Fatal("failed to migrate database: ", err)
	}

	seedData(db)

	userRepo := repository.NewUserRepository(db)
	productRepo := repository.NewProductRepository(db)
	orderRepo := repository.NewOrderRepository(db)
	settlementRepo := repository.NewSettlementRepository(db)

	authService := service.NewAuthService(userRepo)
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

	r := gin.New()
	r.Use(middleware.LoggerMiddleware())
	r.Use(gin.Recovery())
	r.Use(middleware.CORSMiddleware())

	v1 := r.Group("/v1")
	{
		v1.POST("/login", authHandler.Login)

		auth := v1.Group("", middleware.AuthMiddleware())
		{
			auth.GET("/products", productHandler.List)
			auth.POST("/products", middleware.RequireRole("ADMIN"), productHandler.Create)
			auth.PUT("/products/:id", middleware.RequireRole("ADMIN"), productHandler.Update)
			auth.DELETE("/products/:id", middleware.RequireRole("ADMIN"), productHandler.Delete)

			auth.GET("/orders", orderHandler.List)
			auth.GET("/orders/:id", orderHandler.Get)
			auth.POST("/orders", orderHandler.Create)
			auth.POST("/sync/orders", orderHandler.BulkSync)
			auth.PATCH("/orders/:id/void", middleware.RequireRole("ADMIN"), orderHandler.Void)
			auth.GET("/orders/unsettled/summary", orderHandler.UnsettledSummary)
			auth.POST("/orders/settle", middleware.RequireRole("ADMIN"), orderHandler.Settle)

			auth.GET("/settlements", settlementHandler.List)
			auth.GET("/settlements/:id", settlementHandler.Get)
		}
	}

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
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

func seedData(db *gorm.DB) {
	var count int64
	db.Model(&model.User{}).Count(&count)
	if count > 0 {
		return
	}
	log.Println("Seeding data...")
	users := []model.User{
		{Username: "admin", PasswordHash: "$2a$10$uIjrPVsZtsoK01VHa6VC8e0t3O62BpTnF/YomtOLAN0BF087eAah2", Name: "Admin User", Role: "ADMIN"},
		{Username: "cashier", PasswordHash: "$2a$10$7OgCWELW2gl7lL/dAmzFkeJVf540NN4ZboNCJYawE6to/b.Z5s/G2", Name: "Cashier User", Role: "CASHIER"},
	}
	for _, u := range users {
		db.Create(&u)
	}
	products := []model.Product{
		{Name: "Chitato", Price: 11500, Category: "Snacks", ImageURL: "https://images.unsplash.com/photo-1621939514649-28b12e81658b", StoreType: "RETAIL"},
		{Name: "Indomie Goreng", Price: 3500, Category: "Snacks", ImageURL: "https://images.unsplash.com/photo-1612929633738-8fe44f7ec841", StoreType: "RETAIL"},
		{Name: "Teh Botol", Price: 5000, Category: "Beverages", ImageURL: "https://images.unsplash.com/photo-1556679343-c7306c1976bc", StoreType: "RETAIL"},
		{Name: "Sabun Mandi", Price: 8000, Category: "Personal Care", ImageURL: "https://images.unsplash.com/photo-1556228578-0d85b1a4d571", StoreType: "RETAIL"},
		{Name: "Pembersih Lantai", Price: 15000, Category: "Household", ImageURL: "https://images.unsplash.com/photo-1585421514284-efb74c2b69ba", StoreType: "RETAIL"},
		{Name: "Nasi Goreng", Price: 25000, Category: "Food", ImageURL: "https://images.unsplash.com/photo-1512058564366-18510be2db19", StoreType: "FNB"},
		{Name: "Es Teh Manis", Price: 8000, Category: "Beverages", ImageURL: "https://images.unsplash.com/photo-1556679343-c7306c1976bc", StoreType: "FNB"},
		{Name: "Pisang Goreng", Price: 12000, Category: "Desserts", ImageURL: "https://images.unsplash.com/photo-1528975604071-b4dc52a2d18c", StoreType: "FNB"},
		{Name: "Kaos Polos", Price: 75000, Category: "Tops", ImageURL: "https://images.unsplash.com/photo-1521572163474-6864f9cf17ab", StoreType: "OUTFIT"},
		{Name: "Celana Jeans", Price: 250000, Category: "Bottoms", ImageURL: "https://images.unsplash.com/photo-1542272604-787c3835535d", StoreType: "OUTFIT"},
		{Name: "Jaket Hoodie", Price: 185000, Category: "Outerwear", ImageURL: "https://images.unsplash.com/photo-1556821840-3a63f95609a7", StoreType: "OUTFIT"},
	}
	for _, p := range products {
		db.Create(&p)
	}
}
