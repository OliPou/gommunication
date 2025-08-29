package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func BUValidator() gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := c.Param("bu")
		if raw == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing bu"})
			return
		}
		bu := strings.ToLower(strings.TrimSpace(raw))
		c.Set("bu", bu)
		c.Next()
	}
}
