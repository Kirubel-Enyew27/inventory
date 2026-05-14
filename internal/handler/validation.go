package handler

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

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

func AsValidationErrors(err error, out *validator.ValidationErrors) bool {
	if err == nil {
		return false
	}
	if ve, ok := err.(validator.ValidationErrors); ok {
		*out = ve
		return true
	}
	if ve, ok := err.(interface{ Unwrap() error }); ok {
		if inner := ve.Unwrap(); inner != nil {
			if v2, ok2 := inner.(validator.ValidationErrors); ok2 {
				*out = v2
				return true
			}
		}
	}
	return false
}
