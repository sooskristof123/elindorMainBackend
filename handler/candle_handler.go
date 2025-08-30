package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"elindor/handler/response"
	"elindor/service"
)

type CandleHandler struct {
	CandleService service.CandleService
}

func (ch *CandleHandler) CandleEndpoints(router *gin.Engine) {
	router.GET("candles", ch.GetCandles)
	router.GET("candles/:id", ch.GetCandleByID)
}

func NewCandleHandler(candleService service.CandleService) *CandleHandler {
	return &CandleHandler{
		CandleService: candleService,
	}
}

func (ch *CandleHandler) GetCandles(ctx *gin.Context) {
	candles, err := ch.CandleService.GetCandles(ctx)

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

func (ch *CandleHandler) GetCandleByID(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Candle ID is required",
		})
		return
	}

	candle, err := ch.CandleService.GetCandleByID(ctx, id)

	if candle == nil {
		ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{
			"message": "Candle not found",
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

	ctx.IndentedJSON(http.StatusOK, candle)
}
