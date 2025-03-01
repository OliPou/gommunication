package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func ValidateRequest(c *gin.Context, params interface{}) error {
	if err := c.ShouldBindJSON(params); err != nil {
		// Check for type conversion errors
		if unmarshalErr, ok := err.(*json.UnmarshalTypeError); ok {
			RespondError(c, http.StatusBadRequest, fmt.Sprintf(
				"Invalid value for field '%s': expected %s, got %s",
				unmarshalErr.Field,
				unmarshalErr.Type,
				unmarshalErr.Value,
			))
			return err
		}

		// Check for validation errors
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			errorMessages := make([]string, 0)
			for _, fe := range ve {
				switch fe.Tag() {
				case "required":
					errorMessages = append(errorMessages, fmt.Sprintf("Field '%s' is required", fe.Field()))
				default:
					errorMessages = append(errorMessages, fmt.Sprintf("Field '%s' validation failed on '%s'", fe.Field(), fe.Tag()))
				}
			}
			RespondError(c, http.StatusBadRequest, strings.Join(errorMessages, "; "))
			return err
		}

		// Generic JSON syntax errors
		if syntaxErr, ok := err.(*json.SyntaxError); ok {
			RespondError(c, http.StatusBadRequest, fmt.Sprintf("Invalid JSON format in request body at position %d", syntaxErr.Offset))
			return err
		}

		// Fallback for any other errors
		RespondError(c, http.StatusBadRequest, "Invalid request body")
		return err
	}
	return nil
}
