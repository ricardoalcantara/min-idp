package users

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-minstack/go-minstack/web"
	"github.com/ricardoalcantara/min-idp/internal/db"
	"github.com/ricardoalcantara/min-idp/internal/rbac"
	rbac_dto "github.com/ricardoalcantara/min-idp/internal/rbac/dto"

	"github.com/ricardoalcantara/min-idp/internal/session"
	session_dto "github.com/ricardoalcantara/min-idp/internal/session/dto"
	user_dto "github.com/ricardoalcantara/min-idp/internal/users/dto"
)

type UserController struct {
	service    *UserService
	rbacSvc    *rbac.RBACService
	sessionSvc *session.SessionService
}

func NewUserController(
	service *UserService,
	rbacSvc *rbac.RBACService,
	sessionSvc *session.SessionService,
) *UserController {
	return &UserController{service: service, rbacSvc: rbacSvc, sessionSvc: sessionSvc}
}

func pageParams(ctx *gin.Context) (page, pageSize int) {
	page, _ = strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ = strconv.Atoi(ctx.DefaultQuery("page_size", "20"))
	return
}

func (c *UserController) list(ctx *gin.Context) {
	page, pageSize := pageParams(ctx)
	us, total, err := c.service.List(page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, web.NewErrorDto(err))
		return
	}
	dtos := make([]user_dto.UserDto, len(us))
	for i := range us {
		dtos[i] = user_dto.NewUserDto(us[i])
	}
	ctx.JSON(http.StatusOK, user_dto.UsersListDto{
		Data:     dtos,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

func (c *UserController) create(ctx *gin.Context) {
	var input user_dto.CreateUserDto
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, web.NewErrorDto(err))
		return
	}
	u, err := c.service.Create(input.Email, input.Username, input.Name, input.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, web.NewErrorDto(err))
		return
	}
	ctx.JSON(http.StatusCreated, user_dto.NewUserDto(u))
}

func (c *UserController) get(ctx *gin.Context) {
	u, err := c.service.FindByUUID(ctx.Param("id"))
	if err != nil {
		if errors.Is(err, db.ErrEntityNotFound) {
			ctx.JSON(http.StatusNotFound, web.NewErrorDto(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, web.NewErrorDto(err))
		return
	}
	ctx.JSON(http.StatusOK, user_dto.NewUserDto(u))
}

func (c *UserController) update(ctx *gin.Context) {
	var input user_dto.UpdateUserDto
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, web.NewErrorDto(err))
		return
	}
	u, err := c.service.Update(ctx.Param("id"), input.Email, input.Username, input.Name, input.Status)
	if err != nil {
		if errors.Is(err, db.ErrEntityNotFound) {
			ctx.JSON(http.StatusNotFound, web.NewErrorDto(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, web.NewErrorDto(err))
		return
	}
	ctx.JSON(http.StatusOK, user_dto.NewUserDto(u))
}

func (c *UserController) delete(ctx *gin.Context) {
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

func (c *UserController) listRoles(ctx *gin.Context) {
	u, err := c.service.FindByUUID(ctx.Param("id"))
	if err != nil {
		if errors.Is(err, db.ErrEntityNotFound) {
			ctx.JSON(http.StatusNotFound, web.NewErrorDto(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, web.NewErrorDto(err))
		return
	}
	roles, err := c.rbacSvc.GetRolesByUser(u.ID)
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

func (c *UserController) assignRole(ctx *gin.Context) {
	u, err := c.service.FindByUUID(ctx.Param("id"))
	if err != nil {
		if errors.Is(err, db.ErrEntityNotFound) {
			ctx.JSON(http.StatusNotFound, web.NewErrorDto(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, web.NewErrorDto(err))
		return
	}
	var input user_dto.AssignRoleDto
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, web.NewErrorDto(err))
		return
	}
	if err := c.rbacSvc.AssignRoleToUserByUUID(u.ID, input.RoleID); err != nil {
		if errors.Is(err, db.ErrEntityNotFound) {
			ctx.JSON(http.StatusNotFound, web.NewErrorDto(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, web.NewErrorDto(err))
		return
	}
	ctx.Status(http.StatusNoContent)
}

func (c *UserController) removeRole(ctx *gin.Context) {
	u, err := c.service.FindByUUID(ctx.Param("id"))
	if err != nil {
		if errors.Is(err, db.ErrEntityNotFound) {
			ctx.JSON(http.StatusNotFound, web.NewErrorDto(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, web.NewErrorDto(err))
		return
	}
	if err := c.rbacSvc.RemoveRoleFromUserByUUID(u.ID, ctx.Param("roleId")); err != nil {
		if errors.Is(err, db.ErrEntityNotFound) {
			ctx.JSON(http.StatusNotFound, web.NewErrorDto(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, web.NewErrorDto(err))
		return
	}
	ctx.Status(http.StatusNoContent)
}

func (c *UserController) sessions(ctx *gin.Context) {
	u, err := c.service.FindByUUID(ctx.Param("id"))
	if err != nil {
		if errors.Is(err, db.ErrEntityNotFound) {
			ctx.JSON(http.StatusNotFound, web.NewErrorDto(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, web.NewErrorDto(err))
		return
	}
	sessions, err := c.sessionSvc.List(ctx.Request.Context(), u.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, web.NewErrorDto(err))
		return
	}
	dtos := make([]session_dto.SessionDto, len(sessions))
	for i := range sessions {
		dtos[i] = session_dto.NewSessionDto(&sessions[i])
	}
	ctx.JSON(http.StatusOK, dtos)
}

func (c *UserController) resetPassword(ctx *gin.Context) {
	u, err := c.service.FindByUUID(ctx.Param("id"))
	if err != nil {
		if errors.Is(err, db.ErrEntityNotFound) {
			ctx.JSON(http.StatusNotFound, web.NewErrorDto(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, web.NewErrorDto(err))
		return
	}
	var input user_dto.ResetPasswordDto
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, web.NewErrorDto(err))
		return
	}
	if err := c.service.UpdatePassword(u.ID, input.Password); err != nil {
		ctx.JSON(http.StatusInternalServerError, web.NewErrorDto(err))
		return
	}
	ctx.Status(http.StatusNoContent)
}
