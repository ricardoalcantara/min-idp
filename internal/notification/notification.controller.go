package notification

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-minstack/web"
	notification_dto "github.com/ricardoalcantara/min-idp/internal/notification/dto"
)

type NotificationController struct {
	service *NotificationService
}

func NewNotificationController(service *NotificationService) *NotificationController {
	return &NotificationController{service: service}
}

func (c *NotificationController) sendTest(ctx *gin.Context) {
	var input notification_dto.SendTestDto
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, web.NewErrorDto(err))
		return
	}

	err := c.service.SendTest(ctx.Request.Context(), input.To)
	if err != nil {
		if errors.Is(err, errSMTPNotConfigured) {
			ctx.JSON(http.StatusServiceUnavailable, web.NewErrorDto(err))
			return
		}
		ctx.JSON(http.StatusBadGateway, web.NewErrorDto(err))
		return
	}

	ctx.JSON(http.StatusOK, web.NewMessageDto("test email sent"))
}
