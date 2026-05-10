package users

import (
	"github.com/gin-gonic/gin"
	"github.com/ricardoalcantara/min-idp/internal/config"
	"github.com/ricardoalcantara/min-idp/internal/session"
)

func RegisterMeRoutes(r *gin.Engine, c *MeController, sessionSvc *session.SessionService, cfg *config.Config) {
	g := r.Group("/api/me", sessionSvc.Middleware(cfg.SessionCookie), session.RequireSession())
	g.GET("", c.me)
	g.PATCH("", c.update)
	g.GET("/sessions", c.sessions)
	g.DELETE("/sessions", c.revokeAllSessions)
	g.DELETE("/sessions/:id", c.revokeSession)
}
