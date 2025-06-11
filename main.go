package main

import (
	"context"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"elindor/handler"
	"elindor/otel"
	"elindor/repository"
	"elindor/service"
)

func main() {
	ctx := context.Background()

	tp := otel.NewOTelProvider(ctx)
	tp.Tracer("elindor-app").Start(ctx, "main")
	defer tp.Shutdown(ctx)

	conn, err := repository.NewRepository()
	if err != nil {
		panic(err)
	}

	candleService := service.NewCandleService(conn)
	candleHandler := handler.NewCandleHandler(candleService)

	router := gin.Default()
	router.Use(otelgin.Middleware("elindor-app"))
	candleHandler.CandleEndpoints(router)

	err = router.Run(":8080")
	if err != nil {
		return
	}
}
