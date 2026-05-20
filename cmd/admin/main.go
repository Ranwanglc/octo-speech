package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"go.uber.org/zap"

	"github.com/Mininglamp-OSS/octo-speech/internal/admin"
	"github.com/Mininglamp-OSS/octo-speech/internal/adminconfig"
	"github.com/Mininglamp-OSS/octo-speech/internal/migration"
	"github.com/Mininglamp-OSS/octo-speech/internal/store"

	"github.com/gin-gonic/gin"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	cfg := adminconfig.LoadFromEnv(logger)

	if cfg.DBDsn == "" {
		logger.Fatal("SPEECH_DB_DSN is required")
	}
	if cfg.Username == "" {
		logger.Fatal("ADMIN_USERNAME is required")
	}
	if cfg.PasswordHash == "" {
		logger.Fatal("ADMIN_PASSWORD is required")
	}

	db, err := sql.Open("mysql", cfg.DBDsn)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		logger.Fatal("failed to ping database", zap.Error(err))
	}

	db.SetMaxOpenConns(50)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	n, err := migration.Run(db)
	if err != nil {
		logger.Fatal("failed to run migrations", zap.Error(err))
	}
	if n > 0 {
		logger.Info("applied migrations", zap.Int("count", n))
	}

	appStore := store.NewAppStore(db, 60)
	auditStore := store.NewAuditStore(db)

	handler := admin.NewHandler(appStore, auditStore, cfg, db, logger)
	rateLimiter := admin.NewLoginRateLimiter(5, time.Minute, cfg.TrustedProxies)

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/healthz", handler.HealthCheck)
	r.GET("/", handler.ServeIndex)
	r.StaticFS("/static", admin.StaticFS())

	api := r.Group("/api")
	{
		api.POST("/login", rateLimiter.Middleware(), handler.Login)

		protected := api.Group("")
		protected.Use(admin.JWTMiddleware(cfg.JWTSecret))
		{
			protected.POST("/logout", handler.Logout)
			protected.GET("/apps", handler.ListApps)
			protected.POST("/apps", admin.CSRFMiddleware(), handler.CreateApp)
			protected.PUT("/apps/:app_id/status", admin.CSRFMiddleware(), handler.UpdateStatus)
			protected.DELETE("/apps/:app_id", admin.CSRFMiddleware(), handler.DeleteApp)
			protected.POST("/apps/:app_id/reset-key", admin.CSRFMiddleware(), handler.ResetKey)
		}
	}

	addr := fmt.Sprintf(":%d", cfg.Port)
	logger.Info("starting octo-speech-admin", zap.String("addr", addr))

	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server failed", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("server forced to shutdown", zap.Error(err))
	}
}
