package main

import (
	"context"

	"elindor/handler"
	"elindor/otel"
	"elindor/service"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

func main() {
	ctx := context.Background()

	tp := otel.NewOTelProvider(ctx)
	tp.Tracer("elindor-app").Start(ctx, "main")
	defer tp.Shutdown(ctx)

	candleService := service.NewCandleService()
	candleHandler := handler.NewCandleHandler(candleService)

	router := gin.Default()
	router.Use(otelgin.Middleware("elindor-app"))
	candleHandler.CandleEndpoints(router)

	err := router.Run(":8080")
	if err != nil {
		return
	}
}
