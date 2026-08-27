package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/tekkpriest/go-portfolio-api/internal/caching"
	"github.com/tekkpriest/go-portfolio-api/internal/config"
	"github.com/tekkpriest/go-portfolio-api/internal/email"
	"github.com/tekkpriest/go-portfolio-api/internal/handlers"
	"github.com/tekkpriest/go-portfolio-api/internal/middleware"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	if err := godotenv.Load(); err != nil {
		slog.Warn("no .env found")
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("invalid config", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	aboutCache := caching.NewAboutCache(cfg.AboutMDPath)
	aboutCache.Start(ctx)

	projectCache := caching.NewProjectCache(cfg.GitHubToken, cfg.GitHubUserName)
	projectCache.Start(ctx)

	mailService := email.NewMailService(cfg.ResendAPIKey, cfg.EmailFrom, cfg.EmailTo)

	aboutHandler := handlers.NewAboutHandler(aboutCache)
	projectHandler := handlers.NewProjectHandler(projectCache)
	contactHandler := handlers.NewContactHandler(mailService)
	healthHandler := handlers.NewHealthHandler(aboutCache, projectCache, 90*time.Minute)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/aboutme", aboutHandler.GetAbout)
	mux.HandleFunc("GET /api/projects", projectHandler.GetProjects)
	mux.HandleFunc("POST /api/contact", contactHandler.PostContact)
	mux.HandleFunc("GET /api/health", healthHandler.GetHealth)

	cors := middleware.CorsHandler(cfg.AllowedOrigin)

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      cors(mux),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		slog.Info("server is starting on", "port", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server encountered an error", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()

	slog.Info("server shutdown started")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("error during server shutdown", "error", err)
		os.Exit(1)
	}

	slog.Info("server shutdown without errors")
}
