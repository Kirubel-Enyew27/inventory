package main

import (
    "fmt"
    "log"
    "net/http"

    "inventory/internal/config"
    "inventory/internal/db"

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
    _ = gormDB

    r := gin.Default()
    r.GET("/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"status": "ok"})
    })

    port := config.Get("PORT", "8080")
    addr := fmt.Sprintf(":%s", port)
    log.Printf("starting server on %s", addr)
    if err := r.Run(addr); err != nil {
        log.Fatalf("server exited: %v", err)
    }
}
