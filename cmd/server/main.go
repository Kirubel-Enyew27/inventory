package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"inventory/internal/config"
	"inventory/internal/db"
	"inventory/internal/handler"
	"inventory/internal/middleware"
	"inventory/internal/repository"
	"inventory/internal/service"

	"github.com/gin-gonic/gin"
)

func main() {
	if err := config.Load(); err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	dsn := config.Get("DB_DSN", "postgres://postgres:postgres@localhost:5432/inventory?sslmode=disable")
	gormDB, err := db.Connect(dsn)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}

	repo := repository.NewProductRepository(gormDB)
	svc := service.NewProductService(repo)
	h := handler.NewProductHandler(svc)

	r := gin.New()
	r.Use(middleware.RequestLogger(), middleware.Recovery())
	r.GET("/health", func(c *gin.Context) {
		if err := db.Ping(c.Request.Context(), gormDB); err != nil {
			handler.JSONError(c, http.StatusServiceUnavailable, "database unavailable")
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	h.RegisterRoutes(r)

	port := config.Get("PORT", "8080")
	addr := fmt.Sprintf(":%s", port)
	server := &http.Server{
		Addr:              addr,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("starting server on %s", addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server exited: %v", err)
		}
	}()

	<-ctx.Done()
	stop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("server shutdown failed: %v", err)
	}
	log.Println("server stopped")
}
