package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
}

func (ch *HealthHandler) HealthEndpoints(router *gin.Engine) {
	router.GET("health", ch.GetHealth)
}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (ch *HealthHandler) GetHealth(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"message": "Service is running",
	})
}
