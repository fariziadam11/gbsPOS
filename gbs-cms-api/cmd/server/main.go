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

	"github.com/gin-gonic/gin"
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

	if cfg.MigrationsPath != "" {
		if err := database.RunMigrations(cfg.DatabaseURL, cfg.MigrationsPath); err != nil {
			log.Fatal("failed to run migrations: ", err)
		}
	} else {
		if err := database.Migrate(db,
			&model.Setting{},
		); err != nil {
			log.Fatal("failed to migrate database: ", err)
		}
	}

	// seedData(db)

	adRepo := repository.NewAdRepository(db)
	playLogRepo := repository.NewAdPlayLogRepository(db)
	userRepo := repository.NewUserRepository(db)
	settingsRepo := repository.NewSettingsRepository(db)

	authService := service.NewAuthService(userRepo, cfg.JWTSecret, cfg.JWTExpiryHours)
	cmsService := service.NewCMSService(adRepo, playLogRepo, cfg.UploadDir)
	settingsService := service.NewSettingsService(settingsRepo)
	userManagementService := service.NewUserService(userRepo)

	authHandler := handler.NewAuthHandler(authService)
	cmsHandler := handler.NewCMSHandler(cmsService, userManagementService)
	settingsHandler := handler.NewSettingsHandler(settingsService)
	userHandler := handler.NewUserHandler(userManagementService)

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
		}
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
		log.Println("CMS API starting on port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("server failed: ", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down CMS API...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatal("forced shutdown: ", err)
	}
	log.Println("CMS API stopped")
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
	log.Println("CMS seeding data...")
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
