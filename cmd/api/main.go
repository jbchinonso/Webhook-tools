package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"gateway/webhook/internal/commons/logger"
	"gateway/webhook/internal/config"
	"gateway/webhook/internal/database"
	"gateway/webhook/internal/middleware"
	"gateway/webhook/internal/modules/health"
	"gateway/webhook/internal/modules/webhooks"
)

// Deps is the composition root. Every dependency that more than one module
// needs lives here; modules receive what they need explicitly via their own
// Init functions.
type Deps struct {
	Log *slog.Logger
	DB  *database.DB
	Cfg *config.Config
}

func main() {
	if err := run(); err != nil {
		slog.New(slog.NewJSONHandler(os.Stderr, nil)).
			Error("fatal", slog.Any("err", err))
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	log := logger.New(cfg.Log.Level, cfg.Log.Format)
	slog.SetDefault(log)

	db, err := database.New(cfg.DB, log, cfg.Log.Level)
	if err != nil {
		return err
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Error("db_close", slog.Any("err", err))
		}
	}()

	deps := &Deps{
		Log: log,
		DB:  db,
		Cfg: cfg,
	}

	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := buildRouter(deps)

	srv := &http.Server{
		Addr:         cfg.HTTP.Addr,
		Handler:      router,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Info("server_starting", slog.String("addr", cfg.HTTP.Addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
		case err := <-errCh:
			log.Error("server_error", slog.Any("err", err))
			// fall through to cleanup, or return err here
			return err
		case sig := <-quit:
			log.Info("shutdown_signal", slog.String("signal", sig.String()))
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Info("shutdown_http")
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("shutdown_error", slog.Any("err", err))
		return err
	}

	log.Info("server_stopped")
	return nil
}

// buildRouter wires every module onto the Gin engine. This is the only
// function that knows the full set of modules. Each module exposes an Init
// function that receives only what it needs.
func buildRouter(d *Deps) (*gin.Engine) {
	router := gin.New()
	router.Use(
		middleware.RequestID(),
		middleware.Logger(d.Log),
		middleware.Recovery(d.Log),
	)

	api := router.Group("/api/v1")

	// ---- Health (unauthenticated) ----
	healthHandler := health.NewHandler(health.NewService(d.DB))
	api.GET("/healthz", healthHandler.Liveness)
	api.GET("/readyz", healthHandler.Readiness)

	// ---- Webhooks ----
	webhooks.Init(router, api, d.DB)

	return router
}
