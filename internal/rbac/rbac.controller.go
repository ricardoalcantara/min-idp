package rbac

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-minstack/web"
	"github.com/ricardoalcantara/min-idp/internal/db"
	rbac_dto "github.com/ricardoalcantara/min-idp/internal/rbac/dto"
)

type RBACController struct {
	service *RBACService
}

func NewRBACController(service *RBACService) *RBACController {
	return &RBACController{service: service}
}

func (c *RBACController) list(ctx *gin.Context) {
	roles, err := c.service.ListRoles()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, web.NewErrorDto(err))
		return
	}
	dtos := make([]rbac_dto.RoleDto, len(roles))
	for i := range roles {
		dtos[i] = rbac_dto.NewRoleDto(&roles[i])
	}
	ctx.JSON(http.StatusOK, dtos)
}

func (c *RBACController) create(ctx *gin.Context) {
	var input rbac_dto.CreateRoleDto
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, web.NewErrorDto(err))
		return
	}
	role, err := c.service.CreateRole(input.Name, input.Description, false)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, web.NewErrorDto(err))
		return
	}
	ctx.JSON(http.StatusCreated, rbac_dto.NewRoleDto(role))
}

func (c *RBACController) get(ctx *gin.Context) {
	role, err := c.service.FindRoleByUUID(ctx.Param("id"))
	if err != nil {
		if errors.Is(err, db.ErrEntityNotFound) {
			ctx.JSON(http.StatusNotFound, web.NewErrorDto(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, web.NewErrorDto(err))
		return
	}
	ctx.JSON(http.StatusOK, rbac_dto.NewRoleDto(role))
}

func (c *RBACController) update(ctx *gin.Context) {
	role, err := c.service.FindRoleByUUID(ctx.Param("id"))
	if err != nil {
		if errors.Is(err, db.ErrEntityNotFound) {
			ctx.JSON(http.StatusNotFound, web.NewErrorDto(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, web.NewErrorDto(err))
		return
	}
	var input rbac_dto.UpdateRoleDto
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, web.NewErrorDto(err))
		return
	}
	updated, err := c.service.UpdateRole(role.ID, input.Name, input.Description)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, web.NewErrorDto(err))
		return
	}
	ctx.JSON(http.StatusOK, rbac_dto.NewRoleDto(updated))
}

func (c *RBACController) delete(ctx *gin.Context) {
	role, err := c.service.FindRoleByUUID(ctx.Param("id"))
	if err != nil {
		if errors.Is(err, db.ErrEntityNotFound) {
			ctx.JSON(http.StatusNotFound, web.NewErrorDto(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, web.NewErrorDto(err))
		return
	}
	if role.System {
		ctx.JSON(http.StatusForbidden, web.NewErrorDto(errors.New("system roles cannot be deleted")))
		return
	}
	if err := c.service.DeleteRole(role.ID); err != nil {
		ctx.JSON(http.StatusInternalServerError, web.NewErrorDto(err))
		return
	}
	ctx.Status(http.StatusNoContent)
}

func (c *RBACController) listPermissions(ctx *gin.Context) {
	role, err := c.service.FindRoleByUUID(ctx.Param("id"))
	if err != nil {
		if errors.Is(err, db.ErrEntityNotFound) {
			ctx.JSON(http.StatusNotFound, web.NewErrorDto(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, web.NewErrorDto(err))
		return
	}
	perms, err := c.service.GetPermissionsByRole(role.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, web.NewErrorDto(err))
		return
	}
	dtos := make([]rbac_dto.PermissionDto, len(perms))
	for i := range perms {
		dtos[i] = rbac_dto.NewPermissionDto(&perms[i])
	}
	ctx.JSON(http.StatusOK, dtos)
}

func (c *RBACController) assignPermission(ctx *gin.Context) {
	var input rbac_dto.AssignPermissionDto
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, web.NewErrorDto(err))
		return
	}
	if err := c.service.AssignPermissionToRoleByUUID(ctx.Param("id"), input.Name); err != nil {
		if errors.Is(err, db.ErrEntityNotFound) {
			ctx.JSON(http.StatusNotFound, web.NewErrorDto(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, web.NewErrorDto(err))
		return
	}
	ctx.Status(http.StatusNoContent)
}

func (c *RBACController) removePermission(ctx *gin.Context) {
	role, err := c.service.FindRoleByUUID(ctx.Param("id"))
	if err != nil {
		if errors.Is(err, db.ErrEntityNotFound) {
			ctx.JSON(http.StatusNotFound, web.NewErrorDto(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, web.NewErrorDto(err))
		return
	}
	perm, err := c.service.repo.FindPermissionByUUID(ctx.Param("permId"))
	if err != nil {
		if errors.Is(err, db.ErrEntityNotFound) {
			ctx.JSON(http.StatusNotFound, web.NewErrorDto(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, web.NewErrorDto(err))
		return
	}
	if err := c.service.RemovePermissionFromRole(role.ID, perm.ID); err != nil {
		ctx.JSON(http.StatusInternalServerError, web.NewErrorDto(err))
		return
	}
	ctx.Status(http.StatusNoContent)
}
