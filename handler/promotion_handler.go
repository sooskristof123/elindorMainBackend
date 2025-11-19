package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"elindor/handler/response"
	"elindor/service"
)

type getAvailablePromotion struct {
	Email         string `json:"email"`
	PromotionName string `json:"promotion_name"`
}

type PromotionHandler struct {
	PromotionService service.PromotionService
}

func (ph *PromotionHandler) PromotionEndpoints(router *gin.Engine) {
	router.POST("promotions", ph.GetPromotionByName)
}

func NewPromotionHandler(promotionService service.PromotionService) *PromotionHandler {
	return &PromotionHandler{
		PromotionService: promotionService,
	}
}

func (ph *PromotionHandler) GetPromotionByName(ctx *gin.Context) {
	var request getAvailablePromotion
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, errors.New("invalid request body"))
		return
	}

	promotion, err := ph.PromotionService.GetPromotionByName(ctx, request.PromotionName, request.Email)

	if err != nil {
		var e response.InternalServerError
		switch {
		case errors.As(err, &e):
			e.RequestID = ctx.GetString("RequestID")
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, e)
		}

		return
	}

	if promotion == nil {
		ctx.AbortWithStatusJSON(http.StatusNotFound, errors.New("promotion not found"))
		return
	}

	if !promotion.Available {
		ctx.AbortWithStatusJSON(http.StatusConflict, errors.New("promotion already used by this email"))
		return
	}

	ctx.IndentedJSON(http.StatusOK, promotion.Promotion)
}
