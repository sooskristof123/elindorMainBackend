package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"elindor/service"
)

type CandleHandler struct {
	CandleService service.CandleService
}

func (ch *CandleHandler) CandleEndpoints(router *gin.Engine) {
	router.GET("candles", ch.GetCandles)
}

func NewCandleHandler(candleService service.CandleService) *CandleHandler {
	return &CandleHandler{
		CandleService: candleService,
	}
}

func (ch *CandleHandler) GetCandles(ctx *gin.Context) {
	candles, err := ch.CandleService.GetCandles(ctx)

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.IndentedJSON(http.StatusOK, candles)
}
