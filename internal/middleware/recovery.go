package middleware

import (
	"inventory/internal/handler"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, err any) {
		log.Printf("panic recovered: %v", err)
		handler.JSONError(c, http.StatusInternalServerError, "internal server error")
	})
}
