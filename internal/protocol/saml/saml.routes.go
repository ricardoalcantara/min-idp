package saml

import (
	"github.com/gin-gonic/gin"
	"github.com/ricardoalcantara/min-idp/internal/config"
	"github.com/ricardoalcantara/min-idp/internal/keystore"
	"github.com/ricardoalcantara/min-idp/internal/kvstore"
	"github.com/ricardoalcantara/min-idp/internal/session"
)

func RegisterRoutes(
	r *gin.Engine,
	c *SAMLController,
	sessionSvc *session.SessionService,
	ks keystore.KeyStore,
	kv kvstore.KVStore,
	cookieToken *session.CookieTokenService,
	cfg *config.Config,
) {
	cookieMw := sessionSvc.CookieMiddleware(cfg.SessionCookie, ks, kv, cookieToken)

	g := r.Group("/saml")
	g.GET("/metadata", c.metadata)
	g.GET("/sso", cookieMw, c.sso)
	g.POST("/sso", cookieMw, c.sso)
	g.GET("/slo", cookieMw, c.slo)
	g.POST("/slo", cookieMw, c.slo)
}
