package oidc

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.Engine, c *OIDCController) {
	wk := r.Group("/.well-known")
	wk.GET("/openid-configuration", c.discovery)
	wk.GET("/jwks.json", c.jwks)

	oauth2 := r.Group("/oauth2")
	oauth2.GET("/authorize", c.authorize)
	oauth2.POST("/token", c.token)
	oauth2.GET("/userinfo", c.userinfo)
	oauth2.POST("/revoke", c.revoke)
	oauth2.POST("/introspect", c.introspect)
	oauth2.GET("/logout", c.logout)
	oauth2.POST("/logout", c.logout)
}
