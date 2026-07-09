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

	_ "gbs-pos-api/docs" // Swagger docs

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
)

// @title GBS POS API
// @version 1.0
// @description Point of Sale REST API for retail, F&B, and fuel station operations
// @termsOfService http://swagger.io/terms/

// @contact.name GBS Support
// @contact.email support@gbs.local

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token. For Keycloak, obtain token from {KEYCLOAK_BASE_URL}/realms/{KEYCLOAK_REALM}/protocol/openid-connect/token

func main() {
	_ = godotenv.Load("../../.env")
	_ = godotenv.Load(".env")

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
	} else if err := db.AutoMigrate(
		&model.User{},
		&model.Product{},
		&model.Discount{},
		&model.Customer{},
		&model.StockMovement{},
		&model.ProductVariant{},
		&model.HoldSession{},
		&model.CartDisplay{},
		&model.Order{},
		&model.OrderItem{},
		&model.Settlement{},
		&model.FuelPrice{},
		&model.Pump{},
		&model.Nozzle{},
		&model.FuelSale{},
	); err != nil {
		log.Fatal("failed to migrate database: ", err)
	}

	database.Seed(db)

	userRepo := repository.NewUserRepository(db)
	productRepo := repository.NewProductRepository(db)
	discountRepo := repository.NewDiscountRepository(db)
	orderRepo := repository.NewOrderRepository(db)
	settlementRepo := repository.NewSettlementRepository(db)
	customerRepo := repository.NewCustomerRepository(db)
	stockMovementRepo := repository.NewStockMovementRepository(db)
	variantRepo := repository.NewProductVariantRepository(db)
	holdRepo := repository.NewHoldRepository(db)
	cartDisplayRepo := repository.NewCartDisplayRepository(db)
	fuelPriceRepo := repository.NewFuelPriceRepository(db)
	pumpRepo := repository.NewPumpRepository(db)
	nozzleRepo := repository.NewNozzleRepository(db)
	fuelSaleRepo := repository.NewFuelSaleRepository(db)

	dashboardRepo := repository.NewDashboardRepository(db)

	authService := service.NewAuthService(userRepo, cfg.JWTSecret, cfg.JWTExpiryHours)
	productService := service.NewProductService(productRepo, stockMovementRepo)
	discountService := service.NewDiscountService(discountRepo, productRepo)
	productService.SetDiscountService(discountService)
	pricingService := service.NewPricingService(productRepo, discountService)
	customerService := service.NewCustomerService(customerRepo)
	variantService := service.NewProductVariantService(variantRepo)
	holdService := service.NewHoldService(holdRepo)
	cartDisplayService := service.NewCartDisplayService(cartDisplayRepo)
	orderService := service.NewOrderService(
		orderRepo,
		productService,
		customerService,
		variantService,
	)
	settlementService := service.NewSettlementService(orderRepo, settlementRepo)
	dashboardService := service.NewDashboardService(dashboardRepo)
	fuelService := service.NewFuelService(fuelPriceRepo, pumpRepo, nozzleRepo, fuelSaleRepo)

	authHandler := handler.NewAuthHandler(authService)
	productHandler := handler.NewProductHandler(productService)
	discountHandler := handler.NewDiscountHandler(discountService, pricingService)
	orderHandler := handler.NewOrderHandler(orderService, settlementService)
	settlementHandler := handler.NewSettlementHandler(settlementService)
	customerHandler := handler.NewCustomerHandler(customerService)
	dashboardHandler := handler.NewDashboardHandler(dashboardService)
	variantHandler := handler.NewProductVariantHandler(variantService)
	holdHandler := handler.NewHoldHandler(holdService)
	cartDisplayHandler := handler.NewCartDisplayHandler(cartDisplayService)
	fuelHandler := handler.NewFuelHandler(fuelService)

	r := router.Setup(
		cfg,
		router.Handlers{
			Auth:           authHandler,
			Product:        productHandler,
			Discount:       discountHandler,
			Order:          orderHandler,
			Settlement:     settlementHandler,
			Customer:       customerHandler,
			Dashboard:      dashboardHandler,
			ProductVariant: variantHandler,
			Hold:           holdHandler,
			Fuel:           fuelHandler,
			CartDisplay:    cartDisplayHandler,
		},
	)

	// Add Swagger UI
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

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
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("server failed: ", err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatal("forced shutdown: ", err)
	}
}
