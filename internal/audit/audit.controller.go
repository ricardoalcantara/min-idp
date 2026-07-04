package audit

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-minstack/go-minstack/web"
	audit_dto "github.com/ricardoalcantara/min-idp/internal/audit/dto"
	audit_repositories "github.com/ricardoalcantara/min-idp/internal/audit/repositories"
)

type AuditController struct {
	service *AuditService
}

func NewAuditController(service *AuditService) *AuditController {
	return &AuditController{service: service}
}

func (c *AuditController) list(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "20"))

	filter := audit_repositories.AuditFilter{
		Action:     ctx.Query("action"),
		TargetType: ctx.Query("target_type"),
		Result:     ctx.Query("result"),
	}
	if s := ctx.Query("since"); s != "" {
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			filter.Since = t
		}
	}
	if u := ctx.Query("until"); u != "" {
		if t, err := time.Parse(time.RFC3339, u); err == nil {
			filter.Until = t
		}
	}

	events, total, err := c.service.List(filter, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, web.NewErrorDto(err))
		return
	}

	dtos := make([]audit_dto.EventDto, len(events))
	for i := range events {
		dtos[i] = audit_dto.NewEventDto(&events[i])
	}
	ctx.JSON(http.StatusOK, audit_dto.EventsListDto{
		Data:     dtos,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}
