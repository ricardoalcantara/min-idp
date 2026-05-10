package users

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-minstack/web"
	user_dto "github.com/ricardoalcantara/min-idp/internal/users/dto"
	"github.com/ricardoalcantara/min-idp/internal/session"
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
	ctx.JSON(http.StatusOK, sessions)
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
