package oidc

import (
	"github.com/gin-gonic/gin"
	"github.com/ricardoalcantara/min-idp/internal/config"
	"github.com/ricardoalcantara/min-idp/internal/keystore"
	"github.com/ricardoalcantara/min-idp/internal/kvstore"
	"github.com/ricardoalcantara/min-idp/internal/session"
)

func RegisterRoutes(
	r *gin.Engine,
	c *OIDCController,
	sessionSvc *session.SessionService,
	ks keystore.KeyStore,
	kv kvstore.KVStore,
	cookieToken *session.CookieTokenService,
	cfg *config.Config,
) {
	wk := r.Group("/.well-known")
	wk.GET("/openid-configuration", c.discovery)
	wk.GET("/jwks.json", c.jwks)

	oauth2 := r.Group("/oauth2")
	cookieMw := sessionSvc.CookieMiddleware(cfg.SessionCookie, ks, kv, cookieToken)

	oauth2.GET("/authorize", cookieMw, c.authorize)
	oauth2.POST("/token", c.token)
	oauth2.GET("/userinfo", c.userinfo)
	oauth2.POST("/revoke", c.revoke)
	oauth2.POST("/introspect", c.introspect)
	oauth2.GET("/logout", cookieMw, c.logout)
	oauth2.POST("/logout", cookieMw, c.logout)
}
