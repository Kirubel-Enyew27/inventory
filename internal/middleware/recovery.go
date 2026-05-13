package middleware

import (
	"log"
	"net/http"

	"inventory/internal/handler"

	"github.com/gin-gonic/gin"
)

// Recovery returns a middleware that recovers from panics and returns JSON error.
func Recovery() gin.HandlerFunc {
    return gin.CustomRecovery(func(c *gin.Context, err interface{}) {
        // Log the panic
        log.Printf("panic recovered: %v", err)
        handler.JSONError(c, http.StatusInternalServerError, "internal server error")
    })
}
