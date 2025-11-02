package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/zuhrulumam/pm-tool/business/domain"
	"github.com/zuhrulumam/pm-tool/business/usecase"
	"github.com/zuhrulumam/pm-tool/config"
	"github.com/zuhrulumam/pm-tool/handler/api"
	"github.com/zuhrulumam/pm-tool/infra/postgres"
	"github.com/zuhrulumam/pm-tool/infra/redis"
	"github.com/zuhrulumam/pm-tool/pkg/middleware"
	oauthhelper "github.com/zuhrulumam/pm-tool/pkg/oauth"
	"github.com/zuhrulumam/pm-tool/pkg/transaction"
	"github.com/zuhrulumam/pm-tool/pkg/telemetry"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the HTTP server",
	Run:   runServer,
}

func runServer(cmd *cobra.Command, args []string) {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}
	// Initialize logger
	logger := telemetry.InitLogger(cfg.Log.Level, cfg.Log.Format)
	log.Logger = logger
	// Initialize tracer
	tracer, cleanup, err := telemetry.InitTracer(cfg.Telemetry.ServiceName, cfg.Telemetry.OTLPEndpoint)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize tracer")
	}
	defer cleanup()

	// Initialize database
	dbCfg := postgres.Config{
		Host:         cfg.Database.Host,
		Port:         cfg.Database.Port,
		User:         cfg.Database.User,
		Password:     cfg.Database.Password,
		Database:     cfg.Database.Name,
		SSLMode:      cfg.Database.SSLMode,
		MaxIdleConns: cfg.Database.MaxIdleConns,
		MaxOpenConns: cfg.Database.MaxOpenConns,
		MaxLifetime:  cfg.Database.ConnMaxLifetime,
	}
	db, err := postgres.NewConnection(dbCfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer db.Close()
	log.Info().Msg("Connected to database")


	// Initialize transaction manager
	txMgr := transaction.NewManager(db)
	// Initialize Redis with telemetry support
	redisCfg := redis.Config{
		Host:     cfg.Redis.Host,
		Port:     cfg.Redis.Port,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	}
	redisClient, err := redis.NewClient(redisCfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to Redis")
	}
	defer redisClient.Close()
	log.Info().Msg("Connected to Redis")
	// init oauthhelper
	oauthhelp, err := oauthhelper.NewOauthProvider(oauthhelper.Dependencies{
		ClientID:     cfg.Oauth.ClientID,
		ClientSecret: cfg.Oauth.ClientSecret,
		RedirectURL:  cfg.Oauth.RedirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		ProviderType: oauthhelper.ProviderGoogle,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create oauth")
	}

	// Step 1: Initialize domains with DB and external services
	domains := domain.NewDomains(domain.DomainDependencies{
		DB:     db,
		Config: cfg,
		Redis: redisClient,
		Tracer:     tracer,
		Oauth:     oauthhelp,
	})
	log.Info().Msg("Domains initialized")

	// Step 2: Initialize use cases with domains and tracer
	usecases := usecase.NewUsecases(usecase.UsecaseDependencies{
		Domains: domains,
		TxMgr:   txMgr,
		Config: cfg,
		Tracer:  tracer,
	})
	log.Info().Msg("Use cases initialized")

	// Step 3: Initialize handlers
	handlers := api.NewHandlers(usecases, cfg, tracer)
	log.Info().Msg("Handlers initialized")

	// Setup Gin router
	router := setupRouter(cfg, cfg.Telemetry.ServiceName)

	// Step 4: Register all routes (handled by handlers)
	handlers.SetupRoutes(router)
	log.Info().Msg("Routes registered")

	// Start server with graceful shutdown
	startServerWithGracefulShutdown(router, cfg)
}

func setupRouter(cfg *config.Config, serviceName string) *gin.Engine {
	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Apply global middleware
	router.Use(middleware.RequestIDMiddleware())
	router.Use(middleware.LoggingMiddleware())
	router.Use(middleware.RecoveryMiddleware())
	router.Use(middleware.CORSMiddleware(
		cfg.CORS.AllowedOrigins,
		cfg.CORS.AllowedMethods,
		cfg.CORS.AllowedHeaders,
	))
	router.Use(middleware.TracingMiddleware(serviceName))

	return router
}

func startServerWithGracefulShutdown(router *gin.Engine, cfg *config.Config) {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.App.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Info().Int("port", cfg.App.Port).Msg("Starting HTTP server")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info().Msg("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Server exited")
}
