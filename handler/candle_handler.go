package handler

import (
	"elindor/service"
	"github.com/gin-gonic/gin"
	"net/http"
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
	candles := ch.CandleService.GetCandles(ctx)
	ctx.IndentedJSON(http.StatusOK, candles)
}
