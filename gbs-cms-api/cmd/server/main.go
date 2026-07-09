// Package main GBS CMS API
//
// Point of Sale & Content Management System API
//
// This API provides endpoints for managing advertisements, users, settings,
// and cart display for the GBS POS system.
//
//	Schemes: http, https
//	BasePath: /v1
//	Version: 1.0.0
//	Host: localhost:8081
//
//	SecurityDefinitions:
//	BearerAuth:
//	  type: apiKey
//	  name: Authorization
//	  in: header
//	  description: JWT Bearer token. For Keycloak auth, obtain token from:
//	  {KEYCLOAK_BASE_URL}/realms/{KEYCLOAK_REALM}/protocol/openid-connect/token
//
//	Responses:
//	401: Unauthorized - Invalid or missing token
//	403: Forbidden - Insufficient permissions
//	422: ValidationError - Request validation failed
//
// swagger:meta
package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"gbs-cms-api/internal/config"
	"gbs-cms-api/internal/database"
	"gbs-cms-api/internal/handler"
	"gbs-cms-api/internal/model"
	"gbs-cms-api/internal/repository"
	"gbs-cms-api/internal/service"
	"gbs-common/middleware"

	_ "gbs-cms-api/docs" // Swagger docs

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

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

	if err := db.AutoMigrate(
		&model.User{},
		&model.Ad{},
		&model.AdPlayLog{},
		&model.CartDisplay{},
	); err != nil {
		log.Fatal("failed to migrate database: ", err)
	}

	seedData(db)

	adRepo := repository.NewAdRepository(db)
	playLogRepo := repository.NewAdPlayLogRepository(db)
	userRepo := repository.NewUserRepository(db)
	settingsRepo := repository.NewSettingsRepository(db)
	cartDisplayRepo := repository.NewCartDisplayRepository(db)

	authService := service.NewAuthService(userRepo, cfg.JWTSecret, cfg.JWTExpiryHours)
	cmsService := service.NewCMSService(adRepo, playLogRepo, cfg.UploadDir)
	settingsService := service.NewSettingsService(settingsRepo)
	userManagementService := service.NewUserService(userRepo)
	cartDisplayService := service.NewCartDisplayService(cartDisplayRepo)

	authHandler := handler.NewAuthHandler(authService)
	cmsHandler := handler.NewCMSHandler(cmsService, userManagementService)
	settingsHandler := handler.NewSettingsHandler(settingsService)
	userHandler := handler.NewUserHandler(userManagementService)
	displayHandler := handler.NewDisplayHandler(cartDisplayService)

	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.MaxMultipartMemory = 32 << 20 // 32 MB
	r.Use(middleware.LoggerMiddleware())
	r.Use(gin.Recovery())
	r.Use(middleware.CORSMiddleware())

	r.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	// Swagger UI
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1 := r.Group("/v1")
	{
		if !cfg.UseKeycloak() || cfg.EnableDemoAuth {
			v1.POST("/login", authHandler.Login)
		}

		authMiddleware, err := buildAuthMiddleware(cfg)
		if err != nil {
			log.Fatal("failed to build auth middleware: ", err)
		}

		auth := v1.Group("", authMiddleware)
		{
			auth.POST("/ads/upload", middleware.RequireRole("ADMIN"), cmsHandler.UploadAd)
			auth.GET("/ads", middleware.RequireRole("ADMIN"), cmsHandler.ListAds)
			auth.GET("/ads/:id", middleware.RequireRole("ADMIN"), cmsHandler.GetAd)
			auth.PUT("/ads/:id", middleware.RequireRole("ADMIN"), cmsHandler.UpdateAd)
			auth.DELETE("/ads/:id", middleware.RequireRole("ADMIN"), cmsHandler.DeleteAd)
			auth.POST("/ads/:id/toggle", middleware.RequireRole("ADMIN"), cmsHandler.ToggleAd)

			auth.GET("/ads/active", cmsHandler.ActivePlaylist)
			auth.GET("/ads/download/:id", cmsHandler.DownloadAd)
			auth.POST("/ads/:id/play", cmsHandler.LogPlay)

			auth.GET("/settings", middleware.RequireRole("ADMIN"), settingsHandler.GetAll)
			auth.PUT("/settings", middleware.RequireRole("ADMIN"), settingsHandler.Update)

			auth.GET("/users", middleware.RequireRole("ADMIN"), userHandler.List)
			auth.GET("/users/:id", middleware.RequireRole("ADMIN"), userHandler.Get)
			auth.POST("/users", middleware.RequireRole("ADMIN"), userHandler.Create)
			auth.PUT("/users/:id", middleware.RequireRole("ADMIN"), userHandler.Update)
			auth.DELETE("/users/:id", middleware.RequireRole("ADMIN"), userHandler.Delete)

			auth.POST("/display/cart", displayHandler.SaveCartDisplay)
		}

		// Public routes for browser/customer displays
		v1.GET("/display/cart", displayHandler.GetCartDisplay)
	}

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

func buildAuthMiddleware(cfg *config.Config) (gin.HandlerFunc, error) {
	if cfg.UseKeycloak() {
		return middleware.NewCompositeAuthMiddleware(cfg.KeycloakJWKSURL(), cfg.JWTSecret)
	}
	return middleware.NewAuthMiddleware(cfg.JWTSecret), nil
}

func seedData(db *gorm.DB) {
	var count int64
	db.Model(&model.User{}).Count(&count)
	if count > 0 {
		return
	}
	users := []model.User{
		{
			Username:     "admin",
			PasswordHash: "$2a$10$uIjrPVsZtsoK01VHa6VC8e0t3O62BpTnF/YomtOLAN0BF087eAah2",
			Name:         "Admin User",
			Role:         "ADMIN",
		},
		{
			Username:     "cashier",
			PasswordHash: "$2a$10$7OgCWELW2gl7lL/dAmzFkeJVf540NN4ZboNCJYawE6to/b.Z5s/G2",
			Name:         "Cashier User",
			Role:         "CASHIER",
		},
	}
	for _, u := range users {
		db.Create(&u)
	}
}
