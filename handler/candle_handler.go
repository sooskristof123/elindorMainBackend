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
