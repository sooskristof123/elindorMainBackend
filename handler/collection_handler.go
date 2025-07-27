package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"elindor/handler/response"
	"elindor/service"
)

type CollectionHandler struct {
	CollectionService service.CollectionService
}

func (ch *CollectionHandler) CollectionEndpoints(router *gin.Engine) {
	router.GET("collections", ch.GetCollections)
	router.GET("/collections/:name/candles", ch.GetCollectionByName)
}

func NewCollectionHandler(collectionService service.CollectionService) *CollectionHandler {
	return &CollectionHandler{
		CollectionService: collectionService,
	}
}

func (ch *CollectionHandler) GetCollections(ctx *gin.Context) {
	collections, err := ch.CollectionService.GetCollections(ctx)

	if err != nil {
		var e response.InternalServerError
		switch {
		case errors.As(err, &e):
			e.RequestID = ctx.GetString("RequestID")
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, e)
		}

		return
	}

	ctx.IndentedJSON(http.StatusOK, collections)
}

func (ch *CollectionHandler) GetCollectionByName(ctx *gin.Context) {
	name := ctx.Param("name")
	if name == "" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Collection name is required",
		})

		return
	}

	candles, err := ch.CollectionService.GetCollection(ctx, name)

	if candles == nil {
		ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{
			"message": "Collection not found",
		})

		return
	}

	if err != nil {
		var e response.InternalServerError
		switch {
		case errors.As(err, &e):
			e.RequestID = ctx.GetString("RequestID")
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, e)
		}

		return
	}

	ctx.IndentedJSON(http.StatusOK, candles)
}
