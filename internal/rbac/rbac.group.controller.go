package rbac

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-minstack/web"
	"github.com/ricardoalcantara/min-idp/internal/db"
	rbac_dto "github.com/ricardoalcantara/min-idp/internal/rbac/dto"
)

type GroupController struct {
	service *GroupService
}

func NewGroupController(service *GroupService) *GroupController {
	return &GroupController{service: service}
}

func (c *GroupController) list(ctx *gin.Context) {
	groups, err := c.service.List()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, web.NewErrorDto(err))
		return
	}
	dtos := make([]rbac_dto.GroupDto, len(groups))
	for i := range groups {
		dtos[i] = rbac_dto.NewGroupDto(&groups[i])
	}
	ctx.JSON(http.StatusOK, dtos)
}

func (c *GroupController) create(ctx *gin.Context) {
	var input rbac_dto.CreateGroupDto
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, web.NewErrorDto(err))
		return
	}
	g, err := c.service.Create(input.Name, input.Description)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, web.NewErrorDto(err))
		return
	}
	ctx.JSON(http.StatusCreated, rbac_dto.NewGroupDto(g))
}

func (c *GroupController) get(ctx *gin.Context) {
	g, err := c.service.FindByUUID(ctx.Param("id"))
	if err != nil {
		if errors.Is(err, db.ErrEntityNotFound) {
			ctx.JSON(http.StatusNotFound, web.NewErrorDto(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, web.NewErrorDto(err))
		return
	}
	ctx.JSON(http.StatusOK, rbac_dto.NewGroupDto(g))
}

func (c *GroupController) update(ctx *gin.Context) {
	var input rbac_dto.UpdateGroupDto
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, web.NewErrorDto(err))
		return
	}
	g, err := c.service.Update(ctx.Param("id"), input.Name, input.Description)
	if err != nil {
		if errors.Is(err, db.ErrEntityNotFound) {
			ctx.JSON(http.StatusNotFound, web.NewErrorDto(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, web.NewErrorDto(err))
		return
	}
	ctx.JSON(http.StatusOK, rbac_dto.NewGroupDto(g))
}

func (c *GroupController) delete(ctx *gin.Context) {
	if err := c.service.Delete(ctx.Param("id")); err != nil {
		if errors.Is(err, db.ErrEntityNotFound) {
			ctx.JSON(http.StatusNotFound, web.NewErrorDto(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, web.NewErrorDto(err))
		return
	}
	ctx.Status(http.StatusNoContent)
}
