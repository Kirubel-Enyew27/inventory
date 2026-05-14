package handler

import "github.com/gin-gonic/gin"

type APIError struct {
	Error   string   `json:"error"`
	Details []string `json:"details,omitempty"`
}

func JSONError(c *gin.Context, status int, message interface{}) {
	var apiErr APIError
	switch v := message.(type) {
	case string:
		apiErr = APIError{Error: v}
	case error:
		apiErr = APIError{Error: v.Error()}
	default:
		apiErr = APIError{Error: "unexpected error"}
	}
	c.AbortWithStatusJSON(status, apiErr)
}
