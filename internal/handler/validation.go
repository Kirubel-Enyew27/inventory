package handler

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// BindJSON binds and validates JSON input, returning a formatted error on failure.
func BindJSON(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBindJSON(obj); err != nil {
		var verrs validator.ValidationErrors
		if ok := AsValidationErrors(err, &verrs); ok {
			var parts []string
			for _, ve := range verrs {
				parts = append(parts, fmt.Sprintf("%s: %s", ve.Field(), ve.Tag()))
			}
			return fmt.Errorf("%s", strings.Join(parts, ", "))
		}
		return err
	}
	return nil
}

// AsValidationErrors attempts to cast err to validator.ValidationErrors.
func AsValidationErrors(err error, out *validator.ValidationErrors) bool {
	if err == nil {
		return false
	}
	return errors.As(err, out)
}
