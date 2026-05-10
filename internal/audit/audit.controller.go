package audit

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-minstack/web"
)

type AuditController struct {
	service *AuditService
}

func NewAuditController(service *AuditService) *AuditController {
	return &AuditController{service: service}
}

func (c *AuditController) list(ctx *gin.Context) {
	ctx.JSON(http.StatusNotImplemented, web.NewMessageDto("not implemented"))
}
