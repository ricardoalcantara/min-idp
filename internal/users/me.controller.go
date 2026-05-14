package users

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-minstack/web"
	"github.com/ricardoalcantara/min-idp/internal/db"
	session_dto "github.com/ricardoalcantara/min-idp/internal/session/dto"
	"github.com/ricardoalcantara/min-idp/internal/session"
	user_dto "github.com/ricardoalcantara/min-idp/internal/users/dto"
)

type MeController struct {
	service    *UserService
	sessionSvc *session.SessionService
}

func NewMeController(service *UserService, sessionSvc *session.SessionService) *MeController {
	return &MeController{service: service, sessionSvc: sessionSvc}
}

func (c *MeController) me(ctx *gin.Context) {
	claims := session.FromContext(ctx)
	u, err := c.service.FindByID(claims.UserID)
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

func (c *MeController) update(ctx *gin.Context) {
	var input user_dto.UpdateMeDto
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, web.NewErrorDto(err))
		return
	}
	claims := session.FromContext(ctx)
	u, err := c.service.UpdateMe(claims.UserID, input.Email, input.Username, input.Name)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, web.NewErrorDto(err))
		return
	}
	ctx.JSON(http.StatusOK, user_dto.NewUserDto(u))
}

func (c *MeController) sessions(ctx *gin.Context) {
	claims := session.FromContext(ctx)
	sessions, err := c.sessionSvc.List(ctx.Request.Context(), claims.UserID)
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

func (c *MeController) revokeAllSessions(ctx *gin.Context) {
	claims := session.FromContext(ctx)
	if err := c.sessionSvc.RevokeAllExcept(ctx.Request.Context(), claims.UserID, claims.SessionUUID); err != nil {
		ctx.JSON(http.StatusInternalServerError, web.NewErrorDto(err))
		return
	}
	ctx.Status(http.StatusNoContent)
}

func (c *MeController) revokeSession(ctx *gin.Context) {
	claims := session.FromContext(ctx)
	targetUUID := ctx.Param("id")

	target, err := c.sessionSvc.GetByUUID(ctx.Request.Context(), targetUUID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, web.NewErrorDto(errors.New("session not found")))
		return
	}
	if target.UserID != claims.UserID {
		ctx.JSON(http.StatusForbidden, web.NewErrorDto(errors.New("forbidden")))
		return
	}
	if err := c.sessionSvc.Revoke(ctx.Request.Context(), targetUUID); err != nil {
		ctx.JSON(http.StatusInternalServerError, web.NewErrorDto(err))
		return
	}
	ctx.Status(http.StatusNoContent)
}
