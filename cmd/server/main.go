package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/tekkpriest/go-portfolio-api/internal/handlers"
	"github.com/tekkpriest/go-portfolio-api/internal/middleware"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env found while initialising server...")
	}

	handlers.StartAboutCache()
	handlers.StartProjectCache()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/aboutme", handlers.GetHandleAbout)
	mux.HandleFunc("GET /api/projects", handlers.GetHandleProjects)
	mux.HandleFunc("POST /api/contact", handlers.PostHandleContact)

	port := os.Getenv("PORT")
	if port == "" {
		port = "7302"
	}

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      middleware.CorsHandler(mux),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	go func() {
		log.Printf("Server is running on Port %s...", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Error while Server shut down: %v", err)
	}

	log.Println("Server was shutdown gracefully.")
}
