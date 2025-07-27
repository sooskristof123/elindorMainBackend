package middleware

import (
	"log"

	"github.com/gin-gonic/gin"
)

func Logging() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		c.Next()
		status := c.Writer.Status()
		requestID, _ := c.Get("RequestID")

		log.Printf("Path: %s | RequestID: %v | Status: %d", path, requestID, status)
	}
}
