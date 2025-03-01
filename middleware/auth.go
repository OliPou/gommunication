// middleware/auth.go
package middleware

import (
	"net/http"

	"github.com/OliPou/gommunication/auth"
	"github.com/gin-gonic/gin"
)

type AuthedHandler func(*gin.Context, string)

func Auth(handler AuthedHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		consumer, err := auth.GetConsumer(c.Request.Header)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Couldn't find consumer"})
			return
		}
		handler(c, consumer)
	}
}
