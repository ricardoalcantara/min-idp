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
	sess := session.FromContext(ctx)
	u, err := c.service.FindByID(sess.UserID)
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
	ctx.JSON(http.StatusNotImplemented, web.NewErrorDto(nil))
}

func (c *MeController) sessions(ctx *gin.Context) {
	sess := session.FromContext(ctx)
	sessions, err := c.sessionSvc.List(ctx.Request.Context(), sess.UserID)
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
	sess := session.FromContext(ctx)
	if err := c.sessionSvc.RevokeAllExcept(ctx.Request.Context(), sess.UserID, sess.ID); err != nil {
		ctx.JSON(http.StatusInternalServerError, web.NewErrorDto(err))
		return
	}
	ctx.Status(http.StatusNoContent)
}

func (c *MeController) revokeSession(ctx *gin.Context) {
	ctx.JSON(http.StatusNotImplemented, web.NewErrorDto(nil))
}
