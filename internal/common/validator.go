package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
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
				case "sendername":
					errorMessages = append(errorMessages, fmt.Sprintf(
						"Field '%s' must be an alphanumeric sender ID (3–15 chars). Spaces/_/- allowed, but it cannot be only digits.",
						fe.Field(),
					))

				case "e164":
					errorMessages = append(errorMessages, fmt.Sprintf(
						"Field '%s' must be a valid phone number in E.164 format (e.g. +33612345678).",
						fe.Field(),
					))
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

func RegisterValidators() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		_ = v.RegisterValidation("e164", validateE164)
		_ = v.RegisterValidation("sendername", validateSenderName)
	}
}

func validateE164(fl validator.FieldLevel) bool {
	e164 := regexp.MustCompile(`^\+?[1-9]\d{7,14}$`)
	s, _ := fl.Field().Interface().(string)
	return e164.MatchString(s)
}

func validateSenderName(fl validator.FieldLevel) bool {
	senderName := regexp.MustCompile(`^[A-Za-z][A-Za-z0-9 _-]{2,14}$`)
	s, _ := fl.Field().Interface().(string)
	if strings.TrimSpace(s) == "" {
		return false
	}
	if regexp.MustCompile(`^\+?\d+$`).MatchString(s) {
		return false
	}
	return senderName.MatchString(s)
}
