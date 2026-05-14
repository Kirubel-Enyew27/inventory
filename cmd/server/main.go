package main

import (
	"fmt"
	"inventory/internal/config"
	"inventory/internal/db"
	"inventory/internal/handler"
	"inventory/internal/middleware"
	"inventory/internal/repository"
	"inventory/internal/service"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	if err := config.Load(); err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	dsn := config.Get("DB_DSN",
		"postgres://postgres:postgres@localhost:5432/inventory?sslmode=disable")
	gormDB, err := db.Connect(dsn)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}

	repo := repository.NewProductRepository(gormDB)
	svc := service.NewProductService(repo)
	h := handler.NewProductHandler(svc)

	r := gin.Default()
	r.Use(middleware.RequestLogger(), middleware.Recovery())
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"Status": "ok"})
	})
	h.RegisterRoutes(r)

	port := config.Get("PORT", "8080")
	addr := fmt.Sprintf(":%s", port)
	log.Printf("starting server on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server exited: %v", err)
	}
}
