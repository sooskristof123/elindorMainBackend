package main

import (
	"context"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"elindor/handler"
	"elindor/middleware"
	"elindor/otel"
	"elindor/repository"
	"elindor/service"
)

func main() {
	ctx := context.Background()

	tp := otel.NewOTelProvider(ctx)
	tp.Tracer("elindor-app").Start(ctx, "main")
	defer tp.Shutdown(ctx)

	repo, err := repository.NewRepository()
	if err != nil {
		panic(err)
	}

	candleService := service.NewCandleService(repo)
	orderService := service.NewOrderService(repo)
	candleHandler := handler.NewCandleHandler(candleService)
	healthHandler := handler.NewHealthHandler()
	orderHandler := handler.NewOrderHandler(orderService)

	collectionService := service.NewCollectionService(repo)
	collectionHandler := handler.NewCollectionHandler(collectionService)

	router := gin.Default()
	router.Use(otelgin.Middleware("elindor-app"))
	router.Use(middleware.RequestID())
	router.Use(middleware.Logging())
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"https://www.elindorcandle.com", "http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "X-Request-ID"},
		ExposeHeaders:    []string{"Content-Length", "X-Request-ID"},
		AllowCredentials: true,
	}))

	candleHandler.CandleEndpoints(router)
	collectionHandler.CollectionEndpoints(router)
	healthHandler.HealthEndpoints(router)
	orderHandler.OrderEndpoints(router)

	addr := "0.0.0.0:8080"
	if os.Getenv("ENV") != "local" {
		addr = ":443"
		err = router.RunTLS(addr, "/etc/letsencrypt/live/api.elindorcandle.com/fullchain.pem", "/etc/letsencrypt/live/api.elindorcandle.com/privkey.pem")
	} else {
		err = router.Run(addr)
	}
	if err != nil {
		return
	}
	if err != nil {
		return
	}
}
