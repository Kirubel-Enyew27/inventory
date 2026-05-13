package main

import (
	"fmt"
	"log"
	"net/http"

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
    // repository / service / handler wiring
    repo := repository.NewProductRepository(gormDB)
    svc := service.NewProductService(repo)
    h := handler.NewProductHandler(svc)

    r := gin.Default()
    // global middleware
    r.Use(middleware.RequestLogger(), middleware.Recovery())
    r.GET("/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"status": "ok"})
    })
    h.RegisterRoutes(r)

    port := config.Get("PORT", "8080")
    addr := fmt.Sprintf(":%s", port)
    log.Printf("starting server on %s", addr)
    if err := r.Run(addr); err != nil {
        log.Fatalf("server exited: %v", err)
    }
}
