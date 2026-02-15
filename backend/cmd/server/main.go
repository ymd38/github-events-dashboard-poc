package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/sessions"
	"github.com/hirokazuyamada/github-events-dashboard-poc/backend/internal/auth"
	"github.com/hirokazuyamada/github-events-dashboard-poc/backend/internal/config"
	"github.com/hirokazuyamada/github-events-dashboard-poc/backend/internal/handler"
	"github.com/hirokazuyamada/github-events-dashboard-poc/backend/internal/middleware"
	"github.com/hirokazuyamada/github-events-dashboard-poc/backend/internal/model"
	"github.com/hirokazuyamada/github-events-dashboard-poc/backend/internal/repository"
	"github.com/hirokazuyamada/github-events-dashboard-poc/backend/internal/service"
	"github.com/hirokazuyamada/github-events-dashboard-poc/backend/internal/sse"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	db, err := repository.NewDB(cfg.DSN())
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()
	r := chi.NewRouter()
	r.Use(middleware.Recovery)
	r.Use(middleware.Logger)
	r.Use(middleware.CORS(cfg.FrontendURL))
	sseHub := sse.NewHub()
	go sseHub.Run()
	eventRepo := repository.NewEventRepository(db)
	userRepo := repository.NewUserRepository(db)
	eventService := service.NewEventService(eventRepo)
	handler.BroadcastFunc = func(event model.Event) {
		sseHub.Broadcast(event)
	}
	sessionManager := auth.NewSessionManager(cfg.SessionSecret)
	oauthHandler := auth.NewOAuthHandler(cfg.GitHubClientID, cfg.GitHubClientSecret, cfg.FrontendURL, sessionManager, userRepo)
	sessionStore := sessions.NewCookieStore([]byte(cfg.SessionSecret))
	sessionStore.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	healthHandler := handler.NewHealthHandler(db)
	webhookHandler := handler.NewWebhookHandler(cfg.GitHubWebhookSecret, eventService)
	eventsHandler := handler.NewEventsHandler(eventService)
	sseHandler := handler.NewSSEHandler(sseHub)
	r.Get("/api/health", healthHandler.ServeHTTP)
	r.Post("/api/webhook", webhookHandler.ServeHTTP)
	r.Get("/api/auth/login", oauthHandler.Login)
	r.Get("/api/auth/callback", oauthHandler.Callback)
	r.Get("/api/auth/me", oauthHandler.Me)
	r.Post("/api/auth/logout", oauthHandler.Logout)
	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth(sessionStore))
		r.Get("/api/events", eventsHandler.List)
		r.Get("/api/events/{id}", eventsHandler.GetByID)
		r.Get("/api/events/stream", sseHandler.ServeHTTP)
	})
	addr := fmt.Sprintf(":%d", cfg.BackendPort)
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	go func() {
		middleware.LogEvent("info", "server starting", map[string]interface{}{"addr": addr})
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	middleware.LogEvent("info", "server shutting down", nil)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}
	middleware.LogEvent("info", "server stopped", nil)
}
